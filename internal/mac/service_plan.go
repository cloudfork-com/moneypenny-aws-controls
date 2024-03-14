package mac

type ServicePlan struct {
	Service
	StateChanges []*StateChange `json:"state-changes"`
}
type StateChange struct {
	DesiredState string `json:"desired-state"`
	Cron         string `json:"cron"`
	CronSpec     CronSpec
}

func (t *ServicePlan) Validate() error {
	for _, each := range t.StateChanges {
		spec, err := ParseCronSpec(each.Cron)
		if err != nil {
			return err
		}
		each.CronSpec = spec
	}
	return nil
}
