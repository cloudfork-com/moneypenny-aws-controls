package mac

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

var serviceTagName = "moneypenny"

type PlanExecutor struct {
	configFile string
	dryRun     bool
	reporting  bool
	wp         *WeekPlan
	plans      []*ServicePlan
	client     *ecs.Client
}

func NewPlanExecutor(localPlanFilename string) (*PlanExecutor, error) {
	p := &PlanExecutor{configFile: localPlanFilename, wp: new(WeekPlan), dryRun: true, reporting: false}
	p.loadServicePlans()
	p.createClient()
	return p, nil
}

func (p *PlanExecutor) Plan() error {
	slog.SetDefault(slog.With("exec", "PLAN"))
	p.dryRun = true
	return p.exec()
}
func (p *PlanExecutor) Apply() error {
	slog.SetDefault(slog.With("exec", "APPLY"))
	p.dryRun = false
	return p.exec()
}
func (p *PlanExecutor) Report() error {
	slog.SetDefault(slog.With("exec", "REPORT"))

	// collect plans from tagges services
	allServices, err := p.fetchAllServices()
	if err != nil {
		return err
	}

	// check existence
	p.wp.TimePlansDo(func(tp *TimePlan) {
		exitsInCluster := slices.ContainsFunc(allServices, func(existing types.Service) bool {
			return *existing.ServiceArn == tp.ARN
		})
		tp.doesNotExist = !exitsInCluster
	})

	rout, _ := os.Create("schedule.html")
	defer rout.Close()
	slog.Info("write schedule")
	rep := Reporter{}
	if err := rep.WriteOn(p.wp, rout); err != nil {
		slog.Error("schedule report failed", "err", err)
		return err
	}
	return nil
}

func (p *PlanExecutor) loadServicePlans() error {
	if len(p.configFile) == 0 {
		slog.Info("no local service plans")
	} else {
		slog.Info("reading service plans", "file", p.configFile)
		data, err := os.ReadFile(p.configFile)
		if err != nil {
			slog.Error("read fail", "err", err)
			return err
		}
		err = json.Unmarshal(data, &p.plans)
		if err != nil {
			slog.Error("parse fail", "err", err)
			return err
		}
		for _, each := range p.plans {
			if err := each.Validate(); err != nil {
				slog.Error("validate fail", "err", err)
				return err
			}
			slog.Info("adding service plan", "service", each.ARN, "disabled", each.Disabled)
			p.wp.AddServicePlan(*each)
		}
	}
	return nil
}

func (p *PlanExecutor) createClient() error {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("config fail", "err", err)
		return err
	}
	p.client = ecs.NewFromConfig(cfg)
	return nil
}

func (p *PlanExecutor) fetchAllServices() ([]types.Service, error) {
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
		slog.Info("adding service plan", "service", *each.ServiceArn, "crons", input)
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
		p.wp.AddServicePlan(*sp)
	}
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
			slog.Warn("disabled plan", "service", each.ARN)
			continue
		}
		event, ok := p.wp.LastScheduledEventAt(each.Service, now)
		if ok {
			lastStatus := ServiceStatus(p.client, each.Service)
			if lastStatus == Unknown {
				exitsInCluster := slices.ContainsFunc(allServices, func(existing types.Service) bool {
					return *existing.ServiceArn == each.ARN
				})
				if exitsInCluster {
					slog.Info("service has unknown last status, assume it is stopped", "name", each.Service.Name())
					lastStatus = Stopped
				} else {
					slog.Warn("service does not exist, update your configuration", "name", each.Service.Name())
					continue
				}
			}
			isRunning := lastStatus == Running
			if event.DesiredState != Running && isRunning {
				slog.Info("[CHANGE] service is running but must be stopped", "name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue)
				if p.dryRun {
					continue
				}
				if err := StopService(p.client, each.Service); err != nil {
					slog.Error("failed to stop service", "err", err, "name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue)
				}
			} else if event.DesiredState == Running && !isRunning {
				slog.Info("[CHANGE] service must be running", "name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue)
				if p.dryRun {
					continue
				}
				if err := StartService(p.client, each.Service); err != nil {
					slog.Error("failed to start service", "err", err, "name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue)
				}
			} else {
				slog.Info("service is in expected state", "name", each.Service.Name(), "state", event.DesiredState, "crons", each.TagValue)
			}
		}
	}
	return nil
}
