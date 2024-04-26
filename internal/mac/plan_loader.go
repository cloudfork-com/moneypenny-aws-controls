package mac

import (
	"encoding/json"
	"log/slog"
	"os"
)

type PlanLoader struct {
	Plans      []*ServicePlan
	configFile string
}

func NewPlanLoader(configFile string) *PlanLoader {
	return &PlanLoader{
		configFile: configFile,
	}
}

func (p *PlanLoader) LoadServicePlans() error {
	if len(p.configFile) == 0 {
		slog.Info("no local service plans")
		return nil
	} else {
		data, err := os.ReadFile(p.configFile)
		if err != nil {
			slog.Error("read fail", "err", err)
			return err
		}
		err = json.Unmarshal(data, &p.Plans)
		if err != nil {
			slog.Error("parse fail", "err", err)
			return err
		}
		for _, each := range p.Plans {
			slog.Info("validating service plan", "name", each.ARN, "cron", each.TagValue)
			if err := each.Validate(); err != nil {
				slog.Error("validate fail", "err", err)
				return err
			}
		}
	}
	slog.Info("read service plans", "file", p.configFile, "count", len(p.Plans))
	return nil
}
