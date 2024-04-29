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
	rout, _ := os.Create("awscontrols-report.html")
	defer rout.Close()

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
	rout, _ := os.Create("awscontrols-schedule.html")
	defer rout.Close()
	r.WriteOpenHTMLOn(rout)
	fmt.Fprintln(rout, "<h2>Schedule</h2>")
	if err := r.WriteScheduleOn(rout); err != nil {
		return err
	}
	return r.WriteCloseHTMLOn(rout)
}

func (r *Reporter) Status() error {
	rout, _ := os.Create("awscontrols-status.html")
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
	if err := rep.WriteOn(r.executor.weekPlan, w); err != nil {
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
		<button class="controlsaction preferred" type="button" onclick="location.href='?do=plan'" >Plan</button>
		<button class="controlsaction" type="button" onclick="location.href='?do=apply'" >Apply</button>
	</div>
`
	fmt.Fprintln(w, content)
	return nil
}
