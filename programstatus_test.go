package tea

import (
	"bytes"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

func TestCursedRenderer_programStatus(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
	r.start()

	step := func(name string, ps *ProgramStatus) string {
		t.Helper()
		out.Reset()
		v := NewView("hello")
		v.ProgramStatus = ps
		r.render(v)
		if err := r.flush(false); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return out.String()
	}

	working := &ProgramStatus{State: ProgramStateWorking, App: "tea", Message: "Building"}
	if got := step("first", working); !strings.Contains(got, ansi.SetProgramStatus(*working)) {
		t.Fatalf("first frame: missing report in %q", got)
	}

	same := *working
	if got := step("unchanged", &same); strings.Contains(got, "7501") {
		t.Fatalf("unchanged status was written again: %q", got)
	}

	changed := same
	changed.Message = "Linking"
	if got := step("changed", &changed); !strings.Contains(got, ansi.SetProgramStatus(changed)) {
		t.Fatalf("changed status not written: %q", got)
	}

	child := &ProgramStatus{State: ProgramStateBlocked, ID: "job", Kind: ProgramStatusKindQuestion}
	got := step("id change", child)
	if strings.Contains(got, ansi.ClearProgramStatusID("job")) {
		t.Fatalf("new record should not be cleared: %q", got)
	}
	assertInOrder(t, got, ansi.ClearProgramStatus, ansi.SetProgramStatus(*child))

	if got := step("removed", nil); !strings.Contains(got, ansi.ClearProgramStatusID("job")) {
		t.Fatalf("removed status not cleared: %q", got)
	}

	if got := step("invalid", &ProgramStatus{State: ProgramStateIdle, ID: "bad id"}); strings.Contains(got, "7501") {
		t.Fatalf("invalid status was written: %q", got)
	}

	done := &ProgramStatus{State: ProgramStateDone, App: "tea"}
	step("done", done)
	out.Reset()
	if err := r.close(); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); strings.Contains(got, "7501") {
		t.Fatalf("close should leave the status in place, got %q", got)
	}

	// Resuming re-sends the last status.
	out.Reset()
	r.start()
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); !strings.Contains(got, ansi.SetProgramStatus(*done)) {
		t.Fatalf("restart did not restore the status: %q", got)
	}
}

func TestTranslateProgramStatusSupport(t *testing.T) {
	var p Program
	if got := p.translateInputEvent(uv.ProgramStatusSupportEvent{}); got != (ProgramStatusSupportMsg{}) {
		t.Fatalf("got %T, want ProgramStatusSupportMsg", got)
	}
}
