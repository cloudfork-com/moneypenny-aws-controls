package mac

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"time"

	_ "embed"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

var serviceTagName = "moneypenny"

type PlanExecutor struct {
	dryRun   bool
	weekPlan *WeekPlan
	plans    []*ServicePlan
	profile  string
	client   *ecs.Client
}

func NewPlanExecutor(client *ecs.Client, plans []*ServicePlan, profile string) *PlanExecutor {
	wp := new(WeekPlan)
	for _, each := range plans {
		wp.AddServicePlan(*each)
	}
	return &PlanExecutor{weekPlan: wp, dryRun: true, profile: profile, plans: plans, client: client}
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
	return StartService(p.client, Service{ARN: serviceARN}, 1)
}

func (p *PlanExecutor) Stop(serviceARN string) error {
	setLogContext("stop", p.profile)
	p.dryRun = false
	if serviceARN == "" {
		return errors.New("no service ARN was given")
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

func (p *PlanExecutor) exec() error {
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
				clog.Info("service has unknown last status, assume it is stopped")
				lastStatus = Stopped
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
