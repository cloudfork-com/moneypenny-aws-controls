package mac

import (
	"log/slog"
	"slices"
	"time"
)

type WeekPlan struct {
	Plans []*DayPlan `json:"plans"`
}

func (w *WeekPlan) AddServicePlan(p ServicePlan) {
	if p.Disabled {
		slog.Warn("service plan disabled", "service", p.ARN)
		return
	}
	for _, each := range p.StateChanges {
		for _, day := range each.CronSpec.DaysOfWeek {
			w.planOfDay(day).AddStateChange(p.Service, each)
		}
	}
}

func (w *WeekPlan) planOfDay(weekday time.Weekday) *DayPlan {
	for _, each := range w.Plans {
		if each.Weekday == weekday {
			return each
		}
	}
	p := &DayPlan{Weekday: weekday}
	w.Plans = append(w.Plans, p)
	return p
}

// ordered list of statechanges on a particular weekday
func (w *WeekPlan) ScheduleForDay(weekday time.Weekday) (list []TimePlan) {
	var dp *DayPlan
	for _, each := range w.Plans {
		if each.Weekday == weekday {
			dp = each
			break
		}
	}
	if dp == nil {
		return
	}
	for _, each := range dp.Plans {
		list = append(list, each)
	}
	slices.SortFunc(list, func(s1, s2 TimePlan) int {
		return intCompare(s1.Hour*60+s1.Minute, s2.Hour*60+s2.Minute)
	})
	return
}

func intCompare(i, j int) int {
	if i == j {
		return 0
	} else {
		if i < j {
			return -1
		} else {
			return 1
		}
	}
}

func (w WeekPlan) ScheduledEventsOn(day time.Time) []ScheduledEvent {
	events := []ScheduledEvent{}
	wkd := day.Weekday()
	for _, dp := range w.Plans {
		if dp.Weekday == wkd {
			for _, tp := range dp.Plans {
				event := ScheduledEvent{
					Service:      tp.Service,
					DesiredState: tp.DesiredState,
					At:           witHourMinute(day, tp.Hour, tp.Minute),
				}
				events = append(events, event)
			}
		}
	}
	return events
}

func witHourMinute(t time.Time, hour, minute int) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), hour, minute, 0, 0, t.Location())
}

func (w WeekPlan) LastScheduledEventAt(service Service, when time.Time) (ScheduledEvent, bool) {
	wkd := when.Weekday()
	event := ScheduledEvent{}
	for _, dp := range w.Plans {
		for _, tp := range dp.Plans {
			if tp.ARN == service.ARN {
				changeAt := time.Date(when.Year(), when.Month(), when.Day(), tp.Hour, tp.Minute, 0, 0, when.Location())
				changeAt = changeAt.Add(time.Duration(dp.Weekday-wkd) * 24 * time.Hour)
				if changeAt.Before(when) && changeAt.After(event.At) {
					event.At = changeAt
					event.Service = tp.Service
					event.DesiredState = tp.DesiredState
				}
				if changeAt.After(when) {
					return event, true
				}
			}
		}
	}
	return event, !event.At.IsZero()
}
