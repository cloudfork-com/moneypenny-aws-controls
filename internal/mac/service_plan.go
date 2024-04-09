package mac

import "time"

type ServicePlan struct {
	Service
	TagValue     string         `json:"moneypenny"`
	StateChanges []*StateChange `json:"state-changes"`
	Disabled     bool           `json:"disabled"`
}

func (t *ServicePlan) Validate() error {
	if t.TagValue != "" {
		chgs, err := ParseStateChanges(t.TagValue)
		if err != nil {
			return err
		}
		t.StateChanges = chgs
		return nil
	}
	for _, each := range t.StateChanges {
		spec, err := ParseCronSpec(each.Cron)
		if err != nil {
			return err
		}
		each.CronSpec = spec
	}
	return nil
}

func (t *ServicePlan) DesiredCountAt(when time.Time) int {
	minutesToday := when.Hour()*60 + when.Minute()
	desired := 0
	for _, each := range t.StateChanges {
		changeToday := each.CronSpec.Hour*60 + each.CronSpec.Minute
		if changeToday > minutesToday {
			return desired
		}
		desired = each.DesiredCount
	}
	return desired
}
