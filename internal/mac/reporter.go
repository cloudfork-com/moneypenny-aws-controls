package mac

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	_ "embed"
)

type Reporter struct {
	executor *PlanExecutor
}

func NewReporter(exec *PlanExecutor) *Reporter {
	return &Reporter{
		executor: exec,
	}
}

func (r *Reporter) Report() error {
	if err := r.executor.BuildWeekPlan(); err != nil {
		return err
	}
	rout, _ := os.Create(fmt.Sprintf("%s-report.html", r.executor.profile))
	defer rout.Close()

	slog.Info("write report")

	r.WriteOpenHTMLOn(rout)

	r.WriteControlsOn(rout)

	fmt.Fprintln(rout, "<h2>Status</h2>")
	if err := r.WriteStatusOn(rout); err != nil {
		return err
	}
	fmt.Fprintln(rout, "<h2>Schedule</h2>")
	if err := r.WriteScheduleOn(rout); err != nil {
		return err
	}
	return r.WriteCloseHTMLOn(rout)
}

func (r *Reporter) Schedule() error {
	if err := r.executor.BuildWeekPlan(); err != nil {
		return err
	}
	rout, _ := os.Create(fmt.Sprintf("%s-schedule.html", r.executor.profile))
	defer rout.Close()
	r.WriteOpenHTMLOn(rout)
	fmt.Fprintln(rout, "<h2>Schedule</h2>")
	if err := r.WriteScheduleOn(rout); err != nil {
		return err
	}
	return r.WriteCloseHTMLOn(rout)
}

func (r *Reporter) Status() error {
	if err := r.executor.BuildWeekPlan(); err != nil {
		return err
	}
	rout, _ := os.Create(fmt.Sprintf("%s-status.html", r.executor.profile))
	defer rout.Close()
	r.WriteOpenHTMLOn(rout)
	fmt.Fprintln(rout, "<h2>Status</h2>")
	if err := r.WriteStatusOn(rout); err != nil {
		return err
	}
	return r.WriteCloseHTMLOn(rout)
}

//go:embed "assets/open.html"
var openHTML string

func (r *Reporter) WriteOpenHTMLOn(w io.Writer) error {
	if _, err := w.Write([]byte(openHTML)); err != nil {
		return err
	}
	return nil
}

//go:embed "assets/close.html"
var closeHTML string

func (r *Reporter) WriteCloseHTMLOn(w io.Writer) error {
	if _, err := w.Write([]byte(closeHTML)); err != nil {
		return err
	}
	return nil
}

func (r *Reporter) WriteScheduleOn(w io.Writer) error {
	rep := ScheduleWriter{}
	if err := rep.WriteOn(r.executor.profile, r.executor.weekPlan, w); err != nil {
		slog.Error("schedule report failed", "err", err)
		return err
	}
	return nil
}

func (r *Reporter) WriteStatusOn(w io.Writer) error {
	rep := StatusWriter{client: r.executor.client}
	if err := rep.WriteOn(r.executor.plans, w); err != nil {
		slog.Error("status writefailed", "err", err)
		return err
	}
	return nil
}

func (r *Reporter) WriteControlsOn(w io.Writer) error {
	content := `
	<div class="controls">
		<button class="controlsaction" type="button" onclick="location.href='?do=schedule'" >Schedule</button>
		<button class="controlsaction preferred" type="button" onclick="location.href='?do=plan&debug=true'" >Plan</button>
		<button class="controlsaction" type="button" onclick="location.href='?do=apply&debug=true'" >Apply</button>
	</div>
`
	fmt.Fprintln(w, content)
	return nil
}
