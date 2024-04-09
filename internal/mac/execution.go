package mac

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"time"

	_ "embed"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

var serviceTagName = "moneypenny"

type PlanExecutor struct {
	dryRun          bool
	weekPlan        *WeekPlan
	localPlans      []*ServicePlan
	plans           []*ServicePlan
	client          *ecs.Client
	profile         string
	fetchedServices []types.Service
}

func NewPlanExecutor(localPlans []*ServicePlan, profile string) (*PlanExecutor, error) {
	p := &PlanExecutor{weekPlan: new(WeekPlan), dryRun: true, profile: profile, localPlans: localPlans}
	p.applyLocalPlans()
	if err := p.createECSClient(); err != nil {
		return nil, err
	}
	return p, nil
}

func setLogContext(action, profile string) {
	clog := slog.With("x", action)
	if profile != "" {
		clog = clog.With("profile", profile)
	}
	slog.SetDefault(clog)
}

func (p *PlanExecutor) Plan() error {
	setLogContext("plan", p.profile)
	return p.exec()
}
func (p *PlanExecutor) Apply() error {
	setLogContext("apply", p.profile)
	p.dryRun = false
	return p.exec()
}

func (p *PlanExecutor) Start(serviceARN string) error {
	setLogContext("start", p.profile)
	p.dryRun = false
	if serviceARN == "" {
		return errors.New("no service ARN was given")
	}
	if err := p.createECSClient(); err != nil {
		return err
	}
	return StartService(p.client, Service{ARN: serviceARN}, 1)
}

func (p *PlanExecutor) Stop(serviceARN string) error {
	setLogContext("stop", p.profile)
	p.dryRun = false
	if serviceARN == "" {
		return errors.New("no service ARN was given")
	}
	if err := p.createECSClient(); err != nil {
		return err
	}
	return StopService(p.client, Service{ARN: serviceARN})
}

func (p *PlanExecutor) ChangeTaskCount(serviceARN string, countInput string) error {
	setLogContext("change-count", p.profile)
	p.dryRun = false
	if serviceARN == "" {
		return errors.New("no service ARN was given")
	}
	if countInput == "" {
		return errors.New("no count was given")
	}
	count, err := strconv.Atoi(countInput)
	if err != nil {
		return err
	}
	if err := p.createECSClient(); err != nil {
		return err
	}
	return ChangeTaskCountOfService(p.client, Service{ARN: serviceARN}, count)
}

func (p *PlanExecutor) Report() error {
	setLogContext("report", p.profile)
	slog.Info("write report")
	return NewReporter(p).Report()
}

func (p *PlanExecutor) Status() error {
	setLogContext("status", p.profile)
	slog.Info("write status")
	return NewReporter(p).Status()
}

func (p *PlanExecutor) Schedule() error {
	setLogContext("schedule", p.profile)
	slog.Info("write schedule")
	return NewReporter(p).Schedule()
}

func (p *PlanExecutor) BuildWeekPlan() error {
	// collect plans from tagges services
	allServices, err := p.fetchAllServices()
	if err != nil {
		return err
	}
	// check existence
	p.weekPlan.TimePlansDo(func(tp *TimePlan) {
		exitsInCluster := slices.ContainsFunc(allServices, func(existing types.Service) bool {
			return *existing.ServiceArn == tp.ARN
		})
		tp.doesNotExist = !exitsInCluster
	})
	return nil
}

func (p *PlanExecutor) applyLocalPlans() {
	for _, each := range p.localPlans {
		slog.Debug("adding local service plan", "service", each.ARN, "disabled", each.Disabled)
		p.weekPlan.AddServicePlan(*each)
	}
}

func (p *PlanExecutor) createECSClient() error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(p.profile))
	if err != nil {
		slog.Error("config fail", "err", err)
		return err
	}
	p.client = ecs.NewFromConfig(cfg)
	return nil
}

func (p *PlanExecutor) fetchAllServices() ([]types.Service, error) {
	if p.fetchedServices != nil {
		return p.fetchedServices, nil
	}
	allServices, err := AllServices(p.client)
	if err != nil {
		slog.Error("AllServices fail", "err", err)
		return nil, err
	}
	for _, each := range allServices {
		input := TagValue(each, serviceTagName)
		if input == "" {
			continue
		}
		sp := new(ServicePlan)
		sp.ARN = *each.ServiceArn
		slog.Debug("adding service plan", "service", *each.ServiceArn, "crons", input)
		if input == "" {
			slog.Warn("invalid moneypenny tag value", "value", input, "err", err)
			continue
		}
		chgs, err := ParseStateChanges(input)
		if err != nil {
			slog.Warn("invalid moneypenny tag value", "value", input, "err", err)
			continue
		}
		sp.TagValue = input
		sp.StateChanges = chgs
		p.plans = append(p.plans, sp)
		p.weekPlan.AddServicePlan(*sp)
	}
	// cache it
	p.fetchedServices = allServices
	return allServices, nil
}

func (p *PlanExecutor) exec() error {
	allServices, err := p.fetchAllServices()
	if err != nil {
		return err
	}
	now := time.Now().In(userLocation)
	slog.Info("executing", "time", now, "location", os.Getenv("TIME_ZONE"))
	for _, each := range p.plans {
		if each.Disabled {
			slog.Warn("disabled plan", "service", each.ARN)
			continue
		}
		event, ok := p.weekPlan.LastScheduledEventAt(each.Service, now)
		if ok {
			howMany, lastStatus := ServiceStatus(p.client, each.Service)
			clog := slog.With("name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue, "task-count", howMany)

			if lastStatus == Unknown {
				exitsInCluster := slices.ContainsFunc(allServices, func(existing types.Service) bool {
					return *existing.ServiceArn == each.ARN
				})
				if exitsInCluster {
					clog.Info("service has unknown last status, assume it is stopped")
					lastStatus = Stopped
				} else {
					clog.Warn("service does not exist, update your configuration")
					continue
				}
			}
			isRunning := lastStatus == Running
			if event.DesiredState != Running && isRunning {
				clog.Info(">> CHANGE: service is running but must be stopped")
				if p.dryRun {
					continue
				}
				if err := StopService(p.client, each.Service); err != nil {
					clog.Error("failed to stop service", "err", err)
				}
			} else if event.DesiredState == Running && !isRunning {
				clog.Info(">> CHANGE: service must be running")
				if p.dryRun {
					continue
				}
				if err := StartService(p.client, each.Service, event.DesiredCount); err != nil {
					clog.Error("failed to start service", "err", err)
				}
			} else {
				if isRunning && event.DesiredCount != howMany {
					clog.Info(">> CHANGE: service must have different task count", "desired", event.DesiredCount)
					if p.dryRun {
						continue
					}
					if err := ChangeTaskCountOfService(p.client, each.Service, event.DesiredCount); err != nil {
						clog.Error("failed to change task count of service", "err", err)
					}
				} else {
					clog.Info("service is in expected state")
				}
			}
		}
	}
	return nil
}
