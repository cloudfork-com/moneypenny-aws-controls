package mac

import (
	"context"
	"encoding/json"
	"fmt"
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
	weekPlan   *WeekPlan
	plans      []*ServicePlan
	client     *ecs.Client
	profile    string
	clog       *slog.Logger
}

func NewPlanExecutor(localPlanFilename string, profile string) (*PlanExecutor, error) {
	p := &PlanExecutor{configFile: localPlanFilename, weekPlan: new(WeekPlan), dryRun: true, profile: profile, clog: slog.Default()}
	if err := p.loadServicePlans(); err != nil {
		return nil, err
	}
	if err := p.createECSClient(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *PlanExecutor) Plan() error {
	p.clog = slog.With("exec", "PLAN", "profile", p.profile)
	return p.exec()
}
func (p *PlanExecutor) Apply() error {
	p.clog = slog.With("exec", "APPLY", "profile", p.profile)
	p.dryRun = false
	return p.exec()
}
func (p *PlanExecutor) Report() error {
	p.clog = slog.With("exec", "REPORT", "profile", p.profile)
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

	rout, _ := os.Create(fmt.Sprintf("%s-schedule.html", p.profile))
	defer rout.Close()
	p.clog.Info("write schedule")
	rep := Reporter{}
	if err := rep.WriteOn(p.profile, p.weekPlan, rout); err != nil {
		p.clog.Error("schedule report failed", "err", err)
		return err
	}
	return nil
}

func (p *PlanExecutor) loadServicePlans() error {
	if len(p.configFile) == 0 {
		p.clog.Info("no local service plans")
	} else {
		p.clog.Info("reading service plans", "file", p.configFile)
		data, err := os.ReadFile(p.configFile)
		if err != nil {
			p.clog.Error("read fail", "err", err)
			return err
		}
		err = json.Unmarshal(data, &p.plans)
		if err != nil {
			p.clog.Error("parse fail", "err", err)
			return err
		}
		for _, each := range p.plans {
			if err := each.Validate(); err != nil {
				p.clog.Error("validate fail", "err", err)
				return err
			}
			if each.Profile == p.profile {
				p.clog.Info("adding service plan", "service", each.ARN, "disabled", each.Disabled)
				p.weekPlan.AddServicePlan(*each)
			}
		}
	}
	return nil
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
		p.clog.Info("adding service plan", "service", *each.ServiceArn, "crons", input)
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
