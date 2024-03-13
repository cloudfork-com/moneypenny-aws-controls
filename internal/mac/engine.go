package mac

import "time"

func (w WeekPlan) ScheduledEventsOn(day time.Time) []ScheduledEvent {
	events := []ScheduledEvent{}
	wkd := int(day.Weekday())
	for _, dp := range w.Plans {
		if dp.Weekday == wkd {
			for _, tp := range dp.Plans {
				event := ScheduledEvent{
					Task:         tp.Task,
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

func (w WeekPlan) LastScheduledEventAt(task Task, when time.Time) (ScheduledEvent, bool) {
	wkd := int(when.Weekday())
	event := ScheduledEvent{}
	for _, dp := range w.Plans {
		if dp.Weekday == wkd {
			for _, tp := range dp.Plans {
				if tp.Name == task.Name {
					changeAt := time.Date(when.Year(), when.Month(), when.Day(), tp.Hour, tp.Minute, 0, 0, when.Location())
					if changeAt.Before(when) && changeAt.After(event.At) {
						event.At = changeAt
						event.Task = tp.Task
						event.DesiredState = tp.DesiredState
					}
					if changeAt.After(when) {
						return event, true
					}
				}
			}
		}
	}
	return event, !event.At.IsZero()
}
