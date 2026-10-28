package tea

import (
	"bytes"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

type programStatusRenderer struct {
	t   *testing.T
	out bytes.Buffer
	r   *cursedRenderer
}

func newProgramStatusRenderer(t *testing.T) *programStatusRenderer {
	t.Helper()
	pr := &programStatusRenderer{t: t}
	pr.r = newCursedRenderer(&pr.out, []string{"TERM=xterm-256color"}, 80, 24)
	pr.r.start()
	return pr
}

// step renders a frame with ps and returns what was written.
func (pr *programStatusRenderer) step(ps *ProgramStatus) string {
	pr.t.Helper()
	pr.out.Reset()
	v := NewView("hello")
	v.ProgramStatus = ps
	pr.r.render(v)
	if err := pr.r.flush(false); err != nil {
		pr.t.Fatal(err)
	}
	return pr.out.String()
}

// close closes the renderer and returns what was written.
func (pr *programStatusRenderer) close() string {
	pr.t.Helper()
	pr.out.Reset()
	if err := pr.r.close(); err != nil {
		pr.t.Fatal(err)
	}
	return pr.out.String()
}

// restart starts the renderer again and returns what was written.
func (pr *programStatusRenderer) restart() string {
	pr.t.Helper()
	pr.out.Reset()
	pr.r.start()
	if err := pr.r.flush(false); err != nil {
		pr.t.Fatal(err)
	}
	return pr.out.String()
}

func TestCursedRenderer_programStatus(t *testing.T) {
	t.Parallel()
	pr := newProgramStatusRenderer(t)

	working := &ProgramStatus{State: ProgramStateWorking, App: "tea", Message: "Building"}
	if got := pr.step(working); !strings.Contains(got, ansi.SetProgramStatus(working.toANSI())) {
		t.Fatalf("first frame: missing report in %q", got)
	}

	same := *working
	if got := pr.step(&same); strings.Contains(got, "7501") {
		t.Fatalf("unchanged status was written again: %q", got)
	}

	changed := same
	changed.Message = "Linking"
	if got := pr.step(&changed); !strings.Contains(got, ansi.SetProgramStatus(changed.toANSI())) {
		t.Fatalf("changed status not written: %q", got)
	}

	if got := pr.step(nil); !strings.Contains(got, ansi.ClearProgramStatus) {
		t.Fatalf("removed status not cleared: %q", got)
	}
}

func TestCursedRenderer_programStatusInvalid(t *testing.T) {
	t.Parallel()
	pr := newProgramStatusRenderer(t)
	pr.step(&ProgramStatus{State: ProgramStateWorking})
	if got := pr.step(&ProgramStatus{State: "busy"}); strings.Contains(got, "7501") {
		t.Fatalf("invalid status touched the terminal: %q", got)
	}
}

func TestCursedRenderer_programStatusClose(t *testing.T) {
	t.Parallel()

	for _, state := range []ProgramState{ProgramStateDone, ProgramStateError} {
		t.Run(string(state), func(t *testing.T) {
			t.Parallel()
			pr := newProgramStatusRenderer(t)
			ps := &ProgramStatus{State: state, App: "tea"}
			pr.step(ps)
			if got := pr.close(); strings.Contains(got, "7501") {
				t.Fatalf("close should leave a %s status, got %q", state, got)
			}
			if got := pr.restart(); !strings.Contains(got, ansi.SetProgramStatus(ps.toANSI())) {
				t.Fatalf("restart did not restore the status: %q", got)
			}
		})
	}

	for _, state := range []ProgramState{ProgramStateIdle, ProgramStateWorking, ProgramStateBlocked} {
		t.Run(string(state), func(t *testing.T) {
			t.Parallel()
			pr := newProgramStatusRenderer(t)
			ps := &ProgramStatus{State: state, App: "tea"}
			pr.step(ps)
			if got := pr.close(); !strings.Contains(got, ansi.ClearProgramStatus) {
				t.Fatalf("close should clear a %s status, got %q", state, got)
			}
			if got := pr.restart(); !strings.Contains(got, ansi.SetProgramStatus(ps.toANSI())) {
				t.Fatalf("restart did not restore the status: %q", got)
			}
		})
	}
}

func TestTranslateProgramStatusSupport(t *testing.T) {
	var p Program
	if got := p.translateInputEvent(uv.ProgramStatusSupportEvent{}); got != (ProgramStatusSupportMsg{}) {
		t.Fatalf("got %T, want ProgramStatusSupportMsg", got)
	}
}
