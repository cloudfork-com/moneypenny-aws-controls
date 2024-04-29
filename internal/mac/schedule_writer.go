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

func (r ScheduleWriter) WriteOn(wp *WeekPlan, w io.Writer) error {
	tmpl, err := r.scheduleTemplate()
	if err != nil {
		return err
	}
	wd := WeekData{}
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
			td.RowClass = "stopped"
			td.TasksCount = 0
			td.Cron = tp.cron
			if tp.DesiredState == Running {
				td.RowClass = "running"
				td.TasksCount = tp.DesiredCount
			}
			if tp.doesNotExist {
				td.RowClass = "absent"
				td.ServiceName = "MISSING: " + td.ServiceName
			}
			dd.Times = append(dd.Times, td)
		}
		wd.Days = append(wd.Days, dd)
	}
	return tre.New(tmpl.Execute(w, wd), "template exec fail")
}

type WeekData struct {
	Days []DayData
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
	TasksCount  int
	ClusterName string
	Cron        string
	Links       []LinkData
	Savings     string
}
type LinkData struct {
	Href  template.URL
	Title string
}
