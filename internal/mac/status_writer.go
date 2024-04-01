package mac

import (
	"fmt"
	"html/template"
	"io"
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
	now := time.Now().In(userLocation)
	wd := WeekData{StateLabel: "Actual State"}
	dd := DayData{}
	day := now.Weekday()
	dd.DayNumber = int(day)
	dd.Name = day.String()

	for _, each := range plans {
		status := ServiceStatus(r.client, each.Service)
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
			ClusterName: each.ClusterName(),
			Cron:        each.TagValue,
		}
		if status == Stopped {
			link := LinkData{Href: template.URL(fmt.Sprintf("?do=start&service-arn=%s", each.Service.ARN)), Title: "Start service"}
			timeData.Links = append(timeData.Links, link)
		} else {
			link := LinkData{Href: template.URL(fmt.Sprintf("?do=stop&service-arn=%s", each.Service.ARN)), Title: "Stop service"}
			timeData.Links = append(timeData.Links, link)
		}
		link := LinkData{Href: template.URL(each.TagsURL()), Title: "Manage tags"}
		timeData.Links = append(timeData.Links, link)

		dd.Times = append(dd.Times, timeData)
	}
	wd.Days = append(wd.Days, dd)

	return tre.New(tmpl.Execute(w, wd), "template exec fail")
}
