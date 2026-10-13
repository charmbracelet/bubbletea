package tea

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

func regressionFrame(t *testing.T, r *cursedRenderer, c *Cursor) {
	t.Helper()
	v := NewView("cursor")
	v.Cursor = c
	r.render(v)
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
}

// Preserve relinquishes shape/blink control; it does not request a host style.
func TestCursorRegressionPreserveThroughHiddenFrame(t *testing.T) {
	for _, hidden := range []bool{false, true} {
		t.Run(map[bool]string{false: "direct", true: "via_hidden"}[hidden], func(t *testing.T) {
			var out bytes.Buffer
			r := newCursedRenderer(&out, []string{"TERM=xterm"}, 80, 24)
			r.requestCursorStyleCapture()
			r.saveCursorStyle(6)
			c := NewCursor(0, 0)
			c.Shape = CursorUnderline
			regressionFrame(t, r, c)
			if hidden {
				regressionFrame(t, r, nil)
			}
			out.Reset()
			c = NewCursor(0, 0)
			c.PreserveStyle = true
			regressionFrame(t, r, c)
			if strings.Contains(out.String(), " q") {
				t.Fatalf("preserve wrote a cursor style: %q", out.String())
			}
		})
	}
}

func TestCursorRegressionExplicitStyles(t *testing.T) {
	for _, shape := range []CursorShape{CursorBlock, CursorUnderline, CursorBar} {
		for _, blink := range []bool{false, true} {
			var out bytes.Buffer
			r := newCursedRenderer(&out, nil, 80, 24)
			c := NewCursor(0, 0)
			c.PreserveStyle = true
			regressionFrame(t, r, c)
			out.Reset()
			c = NewCursor(0, 0)
			c.Shape, c.Blink = shape, blink
			regressionFrame(t, r, c)
			// Expected protocol values are independent of the renderer encoder.
			want := map[CursorShape][2]int{CursorBlock: {2, 1}, CursorUnderline: {4, 3}, CursorBar: {6, 5}}[shape][0]
			if blink {
				want = map[CursorShape][2]int{CursorBlock: {2, 1}, CursorUnderline: {4, 3}, CursorBar: {6, 5}}[shape][1]
			}
			if lastCursorStyle(out.String(), -1) != want {
				t.Fatalf("preserve to explicit %d/%v: %q", shape, blink, out.String())
			}
		}
	}
}

func TestCursorRegressionPreserveScreenChanges(t *testing.T) {
	var out bytes.Buffer
	r := newCursedRenderer(&out, nil, 80, 24)
	c := NewCursor(0, 0)
	c.PreserveStyle = true
	for _, alt := range []bool{false, true, false} {
		v := NewView("cursor")
		v.Cursor, v.AltScreen = c, alt
		r.render(v)
		r.resize(90, 25)
		r.clearScreen()
		if err := r.flush(false); err != nil {
			t.Fatal(err)
		}
	}
	if strings.Contains(out.String(), " q") {
		t.Fatalf("screen lifecycle changed style: %q", out.String())
	}
}

func TestCursorRegressionPreserveStillRestoresOwnedStyle(t *testing.T) {
	var out bytes.Buffer
	r := newCursedRenderer(&out, nil, 80, 24)
	r.requestCursorStyleCapture()
	r.saveCursorStyle(6)
	regressionFrame(t, r, NewCursor(0, 0))
	c := NewCursor(0, 0)
	c.PreserveStyle = true
	regressionFrame(t, r, c)
	out.Reset()
	if err := r.close(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), ansi.SetCursorStyle(6)) {
		t.Fatal("preserve forgot previous ownership at cleanup")
	}
}

type cursorRegressionModel struct {
	preserve bool
	reports  int
	wait     bool
}

func (m *cursorRegressionModel) Init() Cmd {
	if m.wait {
		return nil
	}
	return Quit
}

func (m *cursorRegressionModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(KeyPressMsg); ok {
		return m, Quit
	}
	if _, ok := msg.(uv.UnknownDcsEvent); ok {
		m.reports++
	}
	return m, nil
}

func (m *cursorRegressionModel) View() View {
	v := NewView("cursor")
	v.Cursor = NewCursor(0, 0)
	v.Cursor.PreserveStyle = m.preserve
	return v
}

func TestCursorRegressionNoQueryForPreservedStyle(t *testing.T) {
	var out bytes.Buffer
	p := NewProgram(&cursorRegressionModel{preserve: true}, WithInput(strings.NewReader("")), WithOutput(&out), WithoutSignals(), WithWindowSize(80, 24))
	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), requestCursorStyle) {
		t.Fatal("queried cursor style although app never owns shape/blink")
	}
}

func TestCursorRegressionNoInputSkipsQuery(t *testing.T) {
	var out bytes.Buffer
	p := NewProgram(&cursorRegressionModel{}, WithInput(nil), WithOutput(&out), WithoutSignals(), WithWindowSize(80, 24))
	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), requestCursorStyle) {
		t.Fatal("query with input disabled")
	}
}

func TestCursorRegressionRendererDisabledSkipsQuery(t *testing.T) {
	var out bytes.Buffer
	p := NewProgram(&cursorRegressionModel{}, WithoutRenderer(), WithInput(strings.NewReader("")), WithOutput(&out), WithoutSignals())
	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), requestCursorStyle) {
		t.Fatal("disabled renderer queried cursor style")
	}
}

var cursorStyleSequence = regexp.MustCompile("\x1b\\[([0-6]) q")

func lastCursorStyle(output string, initial int) int {
	matches := cursorStyleSequence.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return initial
	}
	return int(matches[len(matches)-1][1][0] - '0')
}

type cursorRegressionOutput struct {
	mu sync.Mutex
	bytes.Buffer
}

func (o *cursorRegressionOutput) Write(b []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.Buffer.Write(b)
}

func (o *cursorRegressionOutput) text() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.Buffer.String()
}

type cursorRegressionTerminal struct {
	cursorRegressionOutput
	input      *io.PipeWriter
	once       sync.Once
	report     string
	afterStyle chan struct{}
	styleOnce  sync.Once
}

func (o *cursorRegressionTerminal) Write(b []byte) (int, error) {
	n, err := o.cursorRegressionOutput.Write(b)
	if o.afterStyle != nil && strings.Contains(string(b), ansi.SetCursorStyle(1)) {
		o.styleOnce.Do(func() { close(o.afterStyle) })
	}
	if strings.Contains(string(b), requestCursorStyle) {
		o.once.Do(func() {
			go func() {
				if o.afterStyle != nil {
					<-o.afterStyle
				}
				if o.report != "" {
					_, _ = io.WriteString(o.input, o.report)
				}
				_, _ = io.WriteString(o.input, "q")
			}()
		})
	}
	return n, err
}

func TestCursorRegressionDecoderAndFallback(t *testing.T) {
	for _, tt := range []struct {
		name, report string
		want         int
		late         bool
	}{
		{"supported", "\x1bP1$r6 q\x1b\\", 6, false},
		{"late_after_first_frame", "\x1bP1$r6 q\x1b\\", 6, true},
		{"unsupported", "\x1bP0$r q\x1b\\", 0, false},
		{"malformed", "\x1bP1$r9 q\x1b\\", 0, false},
		{"no_reply", "", 0, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			in, writer := io.Pipe()
			defer in.Close()
			defer writer.Close()
			out := &cursorRegressionTerminal{input: writer, report: tt.report}
			if tt.late {
				out.afterStyle = make(chan struct{})
			}
			p := NewProgram(&cursorRegressionModel{wait: true}, WithContext(t.Context()), WithInput(in), WithOutput(out), WithoutSignals(), WithWindowSize(80, 24))
			if _, err := p.Run(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.text(), ansi.SetCursorStyle(tt.want)) {
				t.Fatalf("missing restored %d: %q", tt.want, out.text())
			}
			if strings.Index(out.text(), requestCursorStyle) > strings.Index(out.text(), ansi.SetCursorStyle(1)) {
				t.Fatal("query after mutation")
			}
		})
	}
}

type cursorRegressionChild struct {
	output  *cursorRegressionOutput
	fail    bool
	sawHost bool
}

func (c *cursorRegressionChild) SetStdin(io.Reader)  {}
func (c *cursorRegressionChild) SetStdout(io.Writer) {}
func (c *cursorRegressionChild) SetStderr(io.Writer) {}
func (c *cursorRegressionChild) Run() error {
	c.sawHost = lastCursorStyle(c.output.text(), -1) == 6
	_, _ = c.output.Write([]byte(ansi.SetCursorStyle(4)))
	_, _ = c.output.Write([]byte("CHILD-END"))
	if c.fail {
		return errors.New("child failed")
	}
	return nil
}

type cursorRegressionExecModel struct {
	child     *cursorRegressionChild
	preserve  bool
	completed bool
	childErr  error
}

func (m *cursorRegressionExecModel) Init() Cmd {
	return Exec(m.child, func(err error) Msg { return execFinishedMsg{err} })
}

func (m *cursorRegressionExecModel) Update(msg Msg) (Model, Cmd) {
	if result, ok := msg.(execFinishedMsg); ok {
		m.completed = true
		m.childErr = result.err
		return m, Quit
	}
	return m, nil
}

func (m *cursorRegressionExecModel) View() View {
	v := NewView("exec")
	v.Cursor = NewCursor(0, 0)
	v.Cursor.PreserveStyle = m.preserve
	return v
}

func TestCursorRegressionExecLifecycle(t *testing.T) {
	for _, preserve := range []bool{false, true} {
		for _, fail := range []bool{false, true} {
			name := map[bool]string{false: "explicit", true: "preserve"}[preserve] + map[bool]string{false: "_success", true: "_failure"}[fail]
			t.Run(name, func(t *testing.T) {
				out := &cursorRegressionOutput{}
				_, _ = out.Write([]byte(ansi.SetCursorStyle(6)))
				child := &cursorRegressionChild{output: out, fail: fail}
				m := &cursorRegressionExecModel{child: child, preserve: preserve}
				p := NewProgram(m, WithInput(strings.NewReader("")), WithOutput(out), WithoutSignals(), WithWindowSize(80, 24))
				r := newCursedRenderer(out, nil, 80, 24)
				r.requestCursorStyleCapture()
				r.saveCursorStyle(6)
				p.renderer = r
				if _, err := p.Run(); err != nil {
					t.Fatal(err)
				}
				if !m.completed || !child.sawHost || (m.childErr != nil) != fail {
					t.Fatalf("exec lifecycle: completed=%v host=%v err=%v", m.completed, child.sawHost, m.childErr)
				}
				want := 6
				if preserve {
					want = 4
				}
				if lastCursorStyle(out.text(), -1) != want {
					t.Fatalf("final style should be %d: %q", want, out.text())
				}
				afterChild := strings.Split(out.text(), "CHILD-END")[1]
				if !preserve && !strings.Contains(afterChild, ansi.SetCursorStyle(1)) {
					t.Fatalf("resume did not reapply app cursor: %q", out.text())
				}
			})
		}
	}
}

type cursorBlockingModel struct{ entered, release chan struct{} }

func (*cursorBlockingModel) Init() Cmd { return nil }
func (m *cursorBlockingModel) Update(msg Msg) (Model, Cmd) {
	if key, ok := msg.(KeyPressMsg); ok && key.String() == "b" {
		close(m.entered)
		<-m.release
		return m, Quit
	}
	return m, nil
}

func (*cursorBlockingModel) View() View {
	v := NewView("blocking")
	v.Cursor = NewCursor(0, 0)
	return v
}

func TestCursorRegressionResponseWhileModelBlocked(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	in, writer := io.Pipe()
	defer in.Close()
	defer writer.Close()
	out := &cursorRegressionOutput{}
	m := &cursorBlockingModel{entered: make(chan struct{}), release: make(chan struct{})}
	p := NewProgram(m, WithContext(ctx), WithInput(in), WithOutput(out), WithoutSignals(), WithWindowSize(80, 24))
	done := make(chan error, 1)
	go func() { _, err := p.Run(); done <- err }()
	processed := make(chan struct{})
	go func() {
		_, _ = io.WriteString(writer, "b")
		<-m.entered
		_, _ = io.WriteString(writer, "\x1bP1$r6 q\x1b\\")
		// Two separately read sentinel keys establish progress past the report
		// while model dispatch remains blocked. No private capture hook is used.
		_, _ = io.WriteString(writer, "x")
		_, _ = io.WriteString(writer, "y")
		close(processed)
	}()
	select {
	case <-processed:
	case <-ctx.Done():
		close(m.release)
		<-done
		t.Fatal("input could not progress past cursor response while model blocked")
	}
	close(m.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if lastCursorStyle(out.text(), -1) != 6 {
		t.Fatalf("response read during blocked model lost at quit: %q", out.text())
	}
}

type cursorQueryModel struct {
	queried bool
	reports int
}

func (*cursorQueryModel) Init() Cmd { return nil }
func (m *cursorQueryModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(uv.UnknownDcsEvent); ok {
		m.reports++
		if !m.queried {
			m.queried = true
			return m, Raw(requestCursorStyle)
		}
		return m, Quit
	}
	return m, nil
}
func (*cursorQueryModel) View() View { v := NewView("query"); v.Cursor = NewCursor(0, 0); return v }

type cursorQueryTerminal struct {
	cursorRegressionOutput
	input io.Writer
	count int
	first string
}

func (o *cursorQueryTerminal) Write(b []byte) (int, error) {
	n, err := o.cursorRegressionOutput.Write(b)
	if strings.Contains(string(b), requestCursorStyle) {
		o.count++
		report := o.first
		if o.count > 1 {
			report = "\x1bP1$r2 q\x1b\\"
		}
		go func() { _, _ = io.WriteString(o.input, report) }()
	}
	return n, err
}

func TestCursorRegressionUnsupportedCaptureThenApplicationQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	in, writer := io.Pipe()
	defer in.Close()
	defer writer.Close()
	out := &cursorQueryTerminal{input: writer, first: "\x1bP0$r q\x1b\\"}
	m := &cursorQueryModel{}
	p := NewProgram(m, WithContext(ctx), WithInput(in), WithOutput(out), WithoutSignals(), WithWindowSize(80, 24))
	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if m.reports != 2 {
		t.Fatalf("application query response swallowed: got %d reports", m.reports)
	}
	if lastCursorStyle(out.text(), -1) != 0 {
		t.Fatalf("application response replaced unknown host snapshot: %q", out.text())
	}
}

type cursorHandoffChild struct{}

func (*cursorHandoffChild) SetStdin(io.Reader)  {}
func (*cursorHandoffChild) SetStdout(io.Writer) {}
func (*cursorHandoffChild) SetStderr(io.Writer) {}
func (c *cursorHandoffChild) Run() error {
	return nil
}

type cursorHandoffModel struct {
	child   *cursorHandoffChild
	reports int
}

func (m *cursorHandoffModel) Init() Cmd {
	return Exec(m.child, func(err error) Msg { return execFinishedMsg{err} })
}

func (m *cursorHandoffModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(execFinishedMsg); ok {
		return m, Raw(requestCursorStyle)
	}
	if _, ok := msg.(uv.UnknownDcsEvent); ok {
		m.reports++
		return m, Quit
	}
	return m, nil
}
func (*cursorHandoffModel) View() View { v := NewView("handoff"); v.Cursor = NewCursor(0, 0); return v }
func TestCursorRegressionPendingCaptureEndsAtHandoff(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	in, writer, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatal(pipeErr)
	}
	defer in.Close()
	defer writer.Close()
	out := &cursorQueryTerminal{input: writer, first: ""}
	m := &cursorHandoffModel{child: &cursorHandoffChild{}}
	p := NewProgram(m, WithContext(ctx), WithInput(in), WithOutput(out), WithoutSignals(), WithWindowSize(80, 24))
	if _, err := p.Run(); err != nil {
		t.Fatalf("%v; reports=%d; terminal=%q", err, m.reports, out.text())
	}
	if m.reports != 1 {
		t.Fatal("reply after handoff was swallowed")
	}
	if lastCursorStyle(out.text(), -1) != 0 {
		t.Fatalf("post-handoff reply became host style: %q", out.text())
	}
}

func TestCursorRegressionFallbackWithoutReport(t *testing.T) {
	var out bytes.Buffer
	r := newCursedRenderer(&out, nil, 80, 24)
	regressionFrame(t, r, NewCursor(0, 0))
	out.Reset()
	if err := r.close(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), ansi.SetCursorStyle(0)) {
		t.Fatalf("missing fallback: %q", out.String())
	}
}

func TestCursorRegressionReleaseRestore(t *testing.T) {
	for _, reset := range []bool{false, true} {
		for _, preserve := range []bool{false, true} {
			t.Run(map[bool]string{false: "release", true: "suspend_reset"}[reset]+map[bool]string{false: "_explicit", true: "_preserve"}[preserve], func(t *testing.T) {
				out := &cursorRegressionOutput{}
				_, _ = out.Write([]byte(ansi.SetCursorStyle(6)))
				r := newCursedRenderer(out, nil, 80, 24)
				r.requestCursorStyleCapture()
				r.saveCursorStyle(6)
				m := &cursorReleaseModel{output: out, preserve: preserve, reset: reset}
				p := NewProgram(m, WithInput(strings.NewReader("")), WithOutput(out), WithoutSignals(), WithWindowSize(80, 24))
				m.program = p
				p.renderer = r
				if _, err := p.Run(); err != nil {
					t.Fatal(err)
				}
				if m.err != nil {
					t.Fatal(m.err)
				}
				want := 1
				if preserve {
					want = 4
				}
				if m.handoff != 6 || m.resumed != want || lastCursorStyle(out.text(), -1) != map[bool]int{false: 6, true: 4}[preserve] {
					t.Fatalf("handoff/resume/final=%d/%d/%d", m.handoff, m.resumed, lastCursorStyle(out.text(), -1))
				}
			})
		}
	}
}

type cursorReleaseModel struct {
	program          *Program
	output           *cursorRegressionOutput
	preserve, reset  bool
	handoff, resumed int
	err              error
}
type cursorReleaseMsg struct{}

func (*cursorReleaseModel) Init() Cmd { return func() Msg { return cursorReleaseMsg{} } }
func (m *cursorReleaseModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(cursorReleaseMsg); !ok {
		return m, nil
	}
	if m.err = m.program.releaseTerminal(m.reset); m.err != nil {
		return m, Quit
	}
	m.handoff = lastCursorStyle(m.output.text(), -1)
	_, _ = m.output.Write([]byte(ansi.SetCursorStyle(4)))
	if m.err = m.program.RestoreTerminal(); m.err != nil {
		return m, Quit
	}
	if m.err = m.program.renderer.flush(false); m.err != nil {
		return m, Quit
	}
	m.resumed = lastCursorStyle(m.output.text(), -1)
	return m, Quit
}

func (m *cursorReleaseModel) View() View {
	v := NewView("release")
	v.Cursor = NewCursor(0, 0)
	v.Cursor.PreserveStyle = m.preserve
	return v
}

type cursorRegressionExit struct {
	p      *Program
	cancel context.CancelFunc
	kind   string
	hidden bool
}
type cursorRegressionExitMsg struct{}

func (m *cursorRegressionExit) Init() Cmd {
	return func() Msg {
		_ = m.p.renderer.flush(false)
		if m.kind == "cancel" {
			m.cancel()
			return nil
		}
		return cursorRegressionExitMsg{}
	}
}

func (m *cursorRegressionExit) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(cursorRegressionExitMsg); ok {
		m.hidden = true
		if m.kind == "panic" {
			panic("cursor cleanup regression")
		}
		if m.kind == "interrupt" {
			return m, Interrupt
		}
		return m, Quit
	}
	return m, nil
}

func (m *cursorRegressionExit) View() View {
	v := NewView("exit")
	if !m.hidden {
		v.Cursor = NewCursor(0, 0)
	}
	return v
}

func TestCursorRegressionExitPaths(t *testing.T) {
	for _, kind := range []string{"quit", "interrupt", "cancel", "panic"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			out := &cursorRegressionOutput{}
			m := &cursorRegressionExit{cancel: cancel, kind: kind}
			p := NewProgram(m, WithContext(ctx), WithInput(strings.NewReader("")), WithOutput(out), WithoutSignals(), WithWindowSize(80, 24))
			m.p = p
			r := newCursedRenderer(out, nil, 80, 24)
			r.requestCursorStyleCapture()
			r.saveCursorStyle(6)
			p.renderer = r
			_, err := p.Run()
			if (err != nil) != (kind != "quit") {
				t.Fatalf("%s unexpected err: %v", kind, err)
			}
			if !strings.Contains(out.text(), ansi.SetCursorStyle(6)) {
				t.Fatalf("%s missing host cleanup: %q", kind, out.text())
			}
		})
	}
}
