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
	wd := WeekData{StateLabel: "Actual State", TasksLabel: "Active Tasks"}
	dd := DayData{}
	day := now.Weekday()
	dd.DayNumber = int(day)
	dd.Name = day.String() + " , " + now.Format(time.RFC3339)

	for _, each := range plans {
		howMany, status := ServiceStatus(r.client, each.Service)
		if status == "UNKNOWN" {
			status = Stopped
		}
		rowClass := "running"
		if status == Stopped {
			rowClass = "stopped"
		}
		timeData := TimeData{
			RowClass:   rowClass,
			TasksCount: howMany,
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
			link := LinkData{
				Href:  template.URL(fmt.Sprintf("?do=stop&service-arn=%s", each.Service.ARN)),
				Title: "Stop service"}
			timeData.Links = append(timeData.Links, link)

		}
		link := LinkData{Href: template.URL(each.TagsURL()), Title: "Manage tags"}
		timeData.Links = append(timeData.Links, link)

		// Up or downscale
		if status == Running {
			// check against desired count
			desired := each.DesiredCountAt(now)
			if desired > howMany {
				link := LinkData{
					Href:  template.URL(fmt.Sprintf("?do=change-count&service-arn=%ss&count=%d", each.Service.ARN, desired)),
					Title: fmt.Sprintf("Upscale (%d) service", desired)}
				timeData.Links = append(timeData.Links, link)
			} else if howMany > 1 {
				link := LinkData{
					Href:  template.URL(fmt.Sprintf("?do=change-count&service-arn=%s&count=1", each.Service.ARN)),
					Title: "Downscale (1) service"}
				timeData.Links = append(timeData.Links, link)
			}
		}
		dd.Times = append(dd.Times, timeData)
	}
	wd.Days = append(wd.Days, dd)

	return tre.New(tmpl.Execute(w, wd), "template exec fail")
}
