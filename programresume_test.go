package tea

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

const testResumeToken = "test-token-0123456789abcdef"

// resumeReports returns the keys of each Program Resume report in out, between
// the time and the mac, after checking it is signed for this process.
func resumeReports(t *testing.T, out string) []string {
	t.Helper()
	var reports []string
	for _, seq := range strings.Split(out, "\x1b]7577;")[1:] {
		seq, _, _ = strings.Cut(seq, "\x07")
		body, mac, _ := strings.Cut(seq, ":mac=")
		if sum := sha256.Sum256([]byte(body + ":" + testResumeToken)); mac != hex.EncodeToString(sum[:]) {
			t.Errorf("bad mac in %q", seq)
		}
		rest, ok := strings.CutPrefix(body, "pid="+strconv.Itoa(os.Getpid())+":time=")
		if !ok {
			t.Errorf("wrong pid in %q", seq)
		}
		_, keys, _ := strings.Cut(rest, ":")
		reports = append(reports, keys)
	}
	return reports
}

// resumeKeys returns the keys of the report for pr.
func resumeKeys(t *testing.T, pr ProgramResume) string {
	t.Helper()
	return resumeReports(t, ansi.SetProgramResume(testResumeToken, os.Getpid(), time.Now(), pr))[0]
}

func TestCursedRenderer_programResume(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color", ansi.ProgramResumeTokenEnv + "=" + testResumeToken}, 80, 24)
	r.start()

	step := func(name string, pr *ProgramResume) []string {
		t.Helper()
		out.Reset()
		v := NewView("hello")
		v.ProgramResume = pr
		r.render(v)
		if _, err := r.flush(false); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return resumeReports(t, out.String())
	}
	expect := func(name string, got []string, want ...string) {
		t.Helper()
		if !slices.Equal(got, want) {
			t.Errorf("%s: got reports %q, want %q", name, got, want)
		}
	}

	nvim := ProgramResume{Args: []string{"nvim", "-S", "/state/1.vim"}, Dir: "/src"}
	expect("first", step("first", &nvim), resumeKeys(t, nvim))
	if !r.reportsProgramResume() {
		t.Error("reportsProgramResume = false while reporting")
	}

	same := ProgramResume{Args: slices.Clone(nvim.Args), Dir: nvim.Dir}
	expect("unchanged", step("unchanged", &same))

	// A model that reuses its slice still gets the change reported.
	args := []string{"app", "1"}
	step("reused slice", &ProgramResume{Args: args})
	args[1] = "2"
	expect("reused slice", step("reused slice changed", &ProgramResume{Args: args}), resumeKeys(t, ProgramResume{Args: []string{"app", "2"}}))

	expect("unreportable", step("unreportable", &ProgramResume{Args: []string{"./app"}}), "clear=1")

	step("set again", &nvim)
	expect("removed", step("removed", nil), "clear=1")
	if r.reportsProgramResume() {
		t.Error("reportsProgramResume = true after removing it")
	}

	// Closing the renderer, as on exit, suspend, or exec, keeps the record.
	step("before close", &nvim)
	out.Reset()
	if err := r.close(); err != nil {
		t.Fatal(err)
	}
	expect("close", resumeReports(t, out.String()))

	// Getting the terminal back reports again.
	out.Reset()
	r.start()
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	expect("restart", resumeReports(t, out.String()), resumeKeys(t, nvim))
}

func TestCursedRenderer_programResumeWithoutToken(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
	r.start()
	v := NewView("hello")
	v.ProgramResume = &ProgramResume{Args: []string{"nvim"}}
	r.render(v)
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "7577") {
		t.Errorf("reported without a token: %q", out.String())
	}
	if r.reportsProgramResume() {
		t.Error("reportsProgramResume = true without a token")
	}
}

type execResumeModel struct {
	cmd    *exec.Cmd
	resume *ProgramResume
}

func (m *execResumeModel) Init() Cmd {
	return ExecProcess(m.cmd, func(err error) Msg { return execFinishedMsg{err} })
}

func (m *execResumeModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(execFinishedMsg); ok {
		return m, Quit
	}
	return m, nil
}

func (m *execResumeModel) View() View {
	v := NewView("hello")
	v.ProgramResume = m.resume
	return v
}

func TestExecProcessProgramResumeToken(t *testing.T) {
	t.Parallel()

	tokenVar := ansi.ProgramResumeTokenEnv + "=" + testResumeToken
	for name, tc := range map[string]struct {
		resume    *ProgramResume
		wantToken bool
	}{
		"reporting":     {&ProgramResume{Args: []string{"app"}}, false},
		"not reporting": {nil, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cmd := successExecCommand()
			cmd.Env = append(os.Environ(), tokenVar)
			p := NewProgram(&execResumeModel{cmd: cmd, resume: tc.resume},
				WithInput(&bytes.Buffer{}),
				WithOutput(&bytes.Buffer{}),
				WithEnvironment([]string{tokenVar}),
			)
			if _, err := p.Run(); err != nil {
				t.Fatal(err)
			}
			if got := slices.Contains(cmd.Env, tokenVar); got != tc.wantToken {
				t.Errorf("child has the token = %v, want %v", got, tc.wantToken)
			}
		})
	}
}
