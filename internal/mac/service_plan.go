package mac

import (
	"slices"
	"time"
)

type ServicePlan struct {
	Service
	TagValue         string         `json:"moneypenny"`
	ResolvedTagValue string         // if TagValue is a reference to another service then this value is the actual tag value with state changes
	StateChanges     []*StateChange `json:"state-changes"` // sorted by time on day
	Disabled         bool           `json:"disabled"`
	TagError         string         `json:"-"`
}

// the actual tag value with state changes
func (t *ServicePlan) tagValue() string {
	if t.ResolvedTagValue != "" {
		return t.ResolvedTagValue
	}
	return t.TagValue
}

func (t *ServicePlan) PercentageRunning() float32 {
	if len(t.StateChanges) == 0 {
		return 1.0 // assume running all the time
	}
	if t.Disabled {
		return 1.0 // assume running all the time
	}
	weekMinutes := 24 * 7 * 60
	dayMinutes := 24 * 60
	runMinutes := 0
	wp := new(WeekPlan)
	wp.AddServicePlan(*t)
	lastState := ""
	var planThatSetLast *TimePlan
	d := 0
	for {
		startMinutesRun := 0
		dp := wp.planOfDay(time.Weekday(d % 7))
		for _, tp := range dp.Plans {
			if tp.DesiredState == Running {
				startMinutesRun = tp.Hour*60 + tp.Minute
				if lastState == Running {
					runMinutes += startMinutesRun
				}
				// did we ran all plans?
				if planThatSetLast == tp {
					goto end
				}
				if lastState == "" {
					planThatSetLast = tp
				}
				lastState = Running
			}
			if tp.DesiredState == Stopped {
				if lastState == Running {
					stopMinutesRun := tp.Hour*60 + tp.Minute
					runMinutes += stopMinutesRun - startMinutesRun
				}
				// did we ran all plans?
				if planThatSetLast == tp {
					if lastState == Running {
						stopMinutesRun := tp.Hour*60 + tp.Minute
						runMinutes += stopMinutesRun - startMinutesRun
					}
					goto end
				}
				if lastState == "" {
					planThatSetLast = tp
				}
				lastState = Stopped
			}
		}
		if lastState == Running {
			runMinutes += dayMinutes - startMinutesRun
		}
		d++
	}
end:
	return float32(runMinutes) / float32(weekMinutes)
}

func (t *ServicePlan) Validate() error {
	changes := t.tagValue()
	if changes == "" {
		return nil
	}
	chgs, err := ParseStateChanges(changes)
	if err != nil {
		t.TagError = "BAD SYNTAX: " + changes
		t.Disabled = true
		return err
	}
	slices.SortFunc(chgs, func(a, b *StateChange) int {
		return intCompare(a.CronSpec.Hour+a.CronSpec.Minute*60, b.CronSpec.Hour+b.CronSpec.Minute*60)
	})
	t.StateChanges = chgs
	return nil
	// this exists when reading from file
	// for _, each := range t.StateChanges {
	// 	spec, err := ParseCronSpec(each.Cron)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	each.CronSpec = spec
	// }
	// return nil
}

func (t *ServicePlan) DesiredCountAt(when time.Time) int {
	minutesToday := when.Hour()*60 + when.Minute()
	weekDay := when.Weekday()
	desired := 0
	for _, each := range t.StateChanges {
		if !each.CronSpec.IsEffectiveOnWeekday(weekDay) {
			continue
		}
		changeToday := each.CronSpec.Hour*60 + each.CronSpec.Minute
		if changeToday > minutesToday {
			return desired
		}
		desired = each.DesiredCount
	}
	return desired
}

func (t *ServicePlan) CronLabel() string {
	if t.TagError != "" {
		return t.TagError
	}
	return t.TagValue
}
