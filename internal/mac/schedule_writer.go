package mac

import (
	"html/template"
	"io"
	"strconv"
	"time"

	_ "embed"

	"github.com/emicklei/tre"
)

//go:embed assets/schedule.html
var scheduleHTML string

type ScheduleWriter struct {
}

func (r ScheduleWriter) scheduleTemplate() (*template.Template, error) {
	tmpl := template.New("schedule")
	tmpl = tmpl.Funcs(template.FuncMap{
		"twoDigits": func(i int) string {
			s := strconv.Itoa(i)
			if len(s) == 1 {
				return "0" + s
			} else {
				return s
			}
		}})
	tmpl, err := tmpl.Parse(scheduleHTML)
	if err != nil {
		return nil, tre.New(err, "parse template fail")
	}
	return tmpl, nil
}

func (r ScheduleWriter) WriteOn(profile string, wp *WeekPlan, w io.Writer) error {
	tmpl, err := r.scheduleTemplate()
	if err != nil {
		return err
	}
	wd := WeekData{Profile: profile, StateLabel: "Desired State"}
	for d := 0; d < 7; d++ {
		dd := DayData{}
		day := time.Weekday(d)
		dd.DayNumber = int(day)
		dd.Name = day.String()
		for _, tp := range wp.ScheduleForDay(day) {
			td := TimeData{}
			td.ClusterName = tp.ClusterName()
			td.ServiceName = tp.Name()
			td.Plan = tp
			td.Cron = "?"
			td.RowClass = "stopped"
			td.Cron = tp.cron
			if tp.DesiredState == Running {
				td.RowClass = "running"
			}
			if tp.doesNotExist {
				td.RowClass = "absent"
				td.ServiceName = "MISSING: " + td.ServiceName
			}
			link := LinkData{Href: template.URL(tp.TagsURL()), Title: "Manage tags"}
			td.Links = append(td.Links, link)
			dd.Times = append(dd.Times, td)
		}
		wd.Days = append(wd.Days, dd)
	}
	return tre.New(tmpl.Execute(w, wd), "template exec fail")
}

type WeekData struct {
	Profile    string
	Days       []DayData
	StateLabel string
}
type DayData struct {
	Name      string
	DayNumber int
	Times     []TimeData
}
type TimeData struct {
	RowClass    string
	Plan        *TimePlan
	ServiceName string
	ClusterName string
	Cron        string
	Links       []LinkData
}
type LinkData struct {
	Href  template.URL
	Title string
}
