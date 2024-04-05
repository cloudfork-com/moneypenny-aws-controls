package mac

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
