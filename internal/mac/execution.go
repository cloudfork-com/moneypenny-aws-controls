package mac

import (
	"context"
	"log/slog"
	"slices"
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
	clog            *slog.Logger
	fetchedServices []types.Service
}

func NewPlanExecutor(localPlans []*ServicePlan, profile string) (*PlanExecutor, error) {
	p := &PlanExecutor{weekPlan: new(WeekPlan), dryRun: true, profile: profile, clog: slog.Default(), localPlans: localPlans}
	p.applyLocalPlans()
	if err := p.createECSClient(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *PlanExecutor) Plan() error {
	p.clog = slog.With("exec", "PLAN")
	if p.profile != "" {
		p.clog = p.clog.With("profile", p.profile)
	}
	return p.exec()
}
func (p *PlanExecutor) Apply() error {
	p.clog = slog.With("exec", "APPLY")
	if p.profile != "" {
		p.clog = p.clog.With("profile", p.profile)
	}
	p.dryRun = false
	return p.exec()
}

func (p *PlanExecutor) Report() error {
	p.clog = slog.With("exec", "REPORT")
	if p.profile != "" {
		p.clog = p.clog.With("profile", p.profile)
	}
	slog.Info("write report")
	return NewReporter(p).Report()
}

func (p *PlanExecutor) Status() error {
	p.clog = slog.With("exec", "STATUS")
	if p.profile != "" {
		p.clog = p.clog.With("profile", p.profile)
	}
	slog.Info("write status")
	return NewReporter(p).Status()
}

func (p *PlanExecutor) Schedule() error {
	p.clog = slog.With("exec", "SCHEDULE")
	if p.profile != "" {
		p.clog = p.clog.With("profile", p.profile)
	}
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
		p.clog.Debug("adding local service plan", "service", each.ARN, "disabled", each.Disabled)
		p.weekPlan.AddServicePlan(*each)
	}
}

func (p *PlanExecutor) createECSClient() error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(p.profile))
	if err != nil {
		p.clog.Error("config fail", "err", err)
		return err
	}
	p.client = ecs.NewFromConfig(cfg)
	return nil
}

func (p *PlanExecutor) fetchAllServices() ([]types.Service, error) {
	if p.fetchedServices != nil {
		return p.fetchedServices, nil
	}
	allServices, err := AllServices(p.clog, p.client)
	if err != nil {
		p.clog.Error("AllServices fail", "err", err)
		return nil, err
	}
	for _, each := range allServices {
		input := TagValue(each, serviceTagName)
		if input == "" {
			continue
		}
		sp := new(ServicePlan)
		sp.ARN = *each.ServiceArn
		p.clog.Debug("adding service plan", "service", *each.ServiceArn, "crons", input)
		if input == "" {
			p.clog.Warn("invalid moneypenny tag value", "value", input, "err", err)
			continue
		}
		chgs, err := ParseStateChanges(input)
		if err != nil {
			p.clog.Warn("invalid moneypenny tag value", "value", input, "err", err)
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
	now := time.Now()
	for _, each := range p.plans {
		if each.Disabled {
			p.clog.Warn("disabled plan", "service", each.ARN)
			continue
		}
		event, ok := p.weekPlan.LastScheduledEventAt(each.Service, now)
		if ok {
			lastStatus := ServiceStatus(p.clog, p.client, each.Service)
			if lastStatus == Unknown {
				exitsInCluster := slices.ContainsFunc(allServices, func(existing types.Service) bool {
					return *existing.ServiceArn == each.ARN
				})
				if exitsInCluster {
					p.clog.Info("service has unknown last status, assume it is stopped", "name", each.Service.Name())
					lastStatus = Stopped
				} else {
					p.clog.Warn("service does not exist, update your configuration", "name", each.Service.Name())
					continue
				}
			}
			isRunning := lastStatus == Running
			if event.DesiredState != Running && isRunning {
				p.clog.Info("[CHANGE] service is running but must be stopped", "name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue)
				if p.dryRun {
					continue
				}
				if err := StopService(p.clog, p.client, each.Service); err != nil {
					p.clog.Error("failed to stop service", "err", err, "name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue)
				}
			} else if event.DesiredState == Running && !isRunning {
				p.clog.Info("[CHANGE] service must be running", "name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue)
				if p.dryRun {
					continue
				}
				if err := StartService(p.clog, p.client, each.Service); err != nil {
					p.clog.Error("failed to start service", "err", err, "name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue)
				}
			} else {
				p.clog.Info("service is in expected state", "name", each.Service.Name(), "state", event.DesiredState, "crons", each.TagValue)
			}
		}
	}
	return nil
}
