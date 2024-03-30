package mac

import (
	"io"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/emicklei/tre"
)

type StatusWriter struct {
	client *ecs.Client
}

func (r *StatusWriter) WriteOn(plans []*ServicePlan, w io.Writer) error {
	tmpl, err := ScheduleWriter{}.scheduleTemplate()
	if err != nil {
		return err
	}
	now := time.Now().In(timeLocation)
	wd := WeekData{StateLabel: "Actual State"}
	dd := DayData{}
	day := now.Weekday()
	dd.DayNumber = int(day)
	dd.Name = day.String()

	for _, each := range plans {
		status := ServiceStatus(slog.Default(), r.client, each.Service)
		if status == "UNKNOWN" {
			status = Stopped
		}
		rowClass := "running"
		if status == Stopped {
			rowClass = "stopped"
		}
		timeData := TimeData{
			RowClass: rowClass,
			Plan: &TimePlan{
				DesiredState: status,
				Hour:         now.Hour(),
				Minute:       now.Minute(),
			},
			ServiceName: each.Name(),
			ClusterARN:  each.ClusterARN(),
			Cron:        each.TagValue,
		}
		dd.Times = append(dd.Times, timeData)
	}
	wd.Days = append(wd.Days, dd)

	return tre.New(tmpl.Execute(w, wd), "template exec fail")
}
