package mac

import (
	"fmt"
	"html/template"
	"io"
	"strconv"
	"time"

	_ "embed"

	"github.com/emicklei/tre"
)

//go:embed report.html
var reportHTML string

type Reporter struct {
}

func (r Reporter) WriteOn(wp *WeekPlan, w io.Writer) error {
	tmpl := template.New("report")
	tmpl = tmpl.Funcs(template.FuncMap{
		"twoDigits": func(i int) string {
			s := strconv.Itoa(i)
			if len(s) == 1 {
				return "0" + s
			} else {
				return s
			}
		}})
	tmpl, err := tmpl.Parse(reportHTML)
	if err != nil {
		return tre.New(err, "parse template fail")
	}
	wd := WeekData{}
	for d := 0; d < 7; d++ {
		dd := DayData{}
		day := time.Weekday(d)
		if now := time.Now(); now.Weekday() == day {
			dd.Today = fmt.Sprintf("is today, %s", now.Format(time.DateOnly))
		}
		dd.DayNumber = int(day)
		dd.Name = day.String()
		for _, tp := range wp.ScheduleForDay(day) {
			td := TimeData{}
			td.ClusterARN = tp.ClusterARN()
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
	Today     string
	Times     []TimeData
}
type TimeData struct {
	RowClass    string
	Plan        *TimePlan
	ServiceName string
	ClusterARN  string
	Cron        string
}
