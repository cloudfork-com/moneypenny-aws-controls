package mac

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/emicklei/tre"
)

//go:embed assets/status.html
var statusHTML string

type StatusWriter struct {
	client *ecs.Client
}

func (r *StatusWriter) statusTemplate() (*template.Template, error) {
	tmpl := template.New("status")
	tmpl = tmpl.Funcs(template.FuncMap{
		"twoDigits": func(i int) string {
			s := strconv.Itoa(i)
			if len(s) == 1 {
				return "0" + s
			} else {
				return s
			}
		}})
	tmpl, err := tmpl.Parse(statusHTML)
	if err != nil {
		return nil, tre.New(err, "parse template fail")
	}
	return tmpl, nil
}

func (r *StatusWriter) WriteOn(plans []*ServicePlan, w io.Writer) error {
	tmpl, err := r.statusTemplate()
	if err != nil {
		return err
	}
	now := time.Now().In(userLocation)
	wd := WeekData{}
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
		if each.Disabled {
			rowClass = "disabled"
		}
		timeData := TimeData{
			RowClass:   rowClass,
			TasksCount: howMany,
			Savings:    fmt.Sprintf("%d%%", 100-int(each.PercentageRunning()*100.0)),
			Plan: &TimePlan{
				DesiredState: status,
				Hour:         now.Hour(),
				Minute:       now.Minute(),
			},
			ServiceName: each.Name(),
			ClusterName: each.ClusterName(),
			Cron:        each.CronLabel(),
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
