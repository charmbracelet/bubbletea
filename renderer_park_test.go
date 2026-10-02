package tea

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/color"
	"io"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

// These tests cover the renderer stopping its ticker while nothing changes.
// Once the ticker is stopped, a change only reaches the output if whatever
// made it woke the renderer, so each test parks the renderer first and then
// checks that the change is drawn.

// safeBuffer is an output that can be read while the program writes to it.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p) //nolint:wrapcheck
}

func (b *safeBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Len()
}

// since returns everything written after the first n bytes.
func (b *safeBuffer) since(n int) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()[n:]
}

// tickRenderer records when the renderer goroutine flushes, which happens
// once per tick of the frame ticker.
type tickRenderer struct {
	renderer

	mu    sync.Mutex
	ticks []time.Time
}

func (r *tickRenderer) flush(closing bool) (bool, error) {
	if !closing {
		r.mu.Lock()
		r.ticks = append(r.ticks, time.Now())
		r.mu.Unlock()
	}
	return r.renderer.flush(closing)
}

// count returns the number of ticks so far.
func (r *tickRenderer) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.ticks)
}

// after returns the ticks after the first n.
func (r *tickRenderer) after(n int) []time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]time.Time(nil), r.ticks[n:]...)
}

type setViewMsg func(*View)

type runCmdMsg struct{ cmd Cmd }

type parkModel struct{ view View }

func (m *parkModel) Init() Cmd { return nil }

func (m *parkModel) Update(msg Msg) (Model, Cmd) {
	switch msg := msg.(type) {
	case setViewMsg:
		msg(&m.view)
	case runCmdMsg:
		return m, msg.cmd
	}
	return m, nil
}

func (m *parkModel) View() View { return m.view }

type parkHarness struct {
	t     *testing.T
	p     *Program
	out   *safeBuffer
	r     *tickRenderer
	frame time.Duration
	done  chan struct{}
	err   error // valid once done is closed
}

func newParkHarness(t *testing.T, fps int, opts ...ProgramOption) *parkHarness {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)

	h := &parkHarness{
		t:     t,
		out:   &safeBuffer{},
		frame: time.Second / time.Duration(fps),
		done:  make(chan struct{}),
	}
	h.r = &tickRenderer{renderer: newCursedRenderer(h.out, nil, 80, 24)}
	opts = append([]ProgramOption{
		WithContext(ctx),
		WithInput(&bytes.Buffer{}),
		WithOutput(h.out),
		WithFPS(fps),
		WithWindowSize(80, 24),
	}, opts...)
	h.p = NewProgram(&parkModel{view: NewView("first")}, opts...)
	h.p.renderer = h.r

	go func() {
		defer close(h.done)
		_, h.err = h.p.Run()
	}()
	t.Cleanup(func() {
		h.p.Kill()
		<-h.done
	})
	return h
}

// waitParked waits until the renderer has stopped ticking.
func (h *parkHarness) waitParked() {
	h.t.Helper()

	// Parked means no tick for well over a frame. The ticker runs for about
	// half a second of unchanged frames before it stops.
	quiet := 6 * h.frame
	deadline := time.Now().Add(10 * time.Second)
	n := h.r.count()
	for time.Now().Before(deadline) {
		time.Sleep(quiet)
		m := h.r.count()
		if m == n && n > 0 {
			return
		}
		n = m
	}
	h.t.Fatal("the renderer never stopped ticking")
}

// waitOutput waits until everything in want is written after mark and
// returns how long that took.
func (h *parkHarness) waitOutput(mark int, within time.Duration, want ...string) time.Duration {
	h.t.Helper()

	start := time.Now()
	for {
		got := h.out.since(mark)
		missing := ""
		for _, w := range want {
			if !strings.Contains(got, w) {
				missing = w
				break
			}
		}
		if missing == "" {
			return time.Since(start)
		}
		if time.Since(start) > within {
			h.t.Fatalf("%q not written within %v after waking a parked renderer; got %q", missing, within, got)
		}
		time.Sleep(time.Millisecond)
	}
}

// waitDrawn waits until the renderer has drawn a frame with the given
// content. Unlike waitOutput, it doesn't depend on how the renderer encodes
// the change.
func (h *parkHarness) waitDrawn(content string, within time.Duration) {
	h.t.Helper()

	r := h.r.renderer.(*cursedRenderer)
	start := time.Now()
	for {
		r.mu.Lock()
		drawn := r.lastView != nil && r.lastView.Content == content
		r.mu.Unlock()
		if drawn {
			return
		}
		if time.Since(start) > within {
			h.t.Fatalf("%q not drawn within %v after waking a parked renderer", content, within)
		}
		time.Sleep(time.Millisecond)
	}
}

func (h *parkHarness) setView(fn func(*View)) {
	h.p.Send(setViewMsg(fn))
}

// waitDone waits for Run to return.
func (h *parkHarness) waitDone() error {
	h.t.Helper()

	const within = 2 * time.Second
	select {
	case <-h.done:
		return h.err
	case <-time.After(within):
		h.t.Fatalf("Run did not return within %v", within)
		return nil
	}
}

// nopExecCommand is an ExecCommand that does nothing.
type nopExecCommand struct{}

func (nopExecCommand) Run() error          { return nil }
func (nopExecCommand) SetStdin(io.Reader)  {}
func (nopExecCommand) SetStdout(io.Writer) {}
func (nopExecCommand) SetStderr(io.Writer) {}

func TestRendererStopsTickingWhenIdle(t *testing.T) {
	t.Parallel()

	h := newParkHarness(t, 60)
	h.waitOutput(0, 5*time.Second, "first")
	h.waitParked()

	n := h.r.count()
	time.Sleep(time.Second)
	if ticks := h.r.count() - n; ticks != 0 {
		t.Errorf("an idle program ticked %d times in a second, want 0", ticks)
	}
}

// Messages that leave the view unchanged wake the renderer, but must not keep
// it ticking.
func TestRendererParksBetweenMessagesThatChangeNothing(t *testing.T) {
	t.Parallel()

	const fps = 60
	h := newParkHarness(t, fps)
	h.waitParked()

	n := h.r.count()
	for range 10 {
		h.setView(func(*View) {})
		time.Sleep(200 * time.Millisecond)
	}
	// Each message costs at most a tick or two. Without parking this is
	// 2s * 60fps = 120 ticks.
	if ticks := h.r.count() - n; ticks > 30 {
		t.Errorf("10 messages that changed nothing caused %d ticks, want at most 30", ticks)
	}
}

// While the view keeps changing, the ticker runs at the frame rate exactly as
// it always has.
func TestRendererTicksAtFrameRateWhileActive(t *testing.T) {
	t.Parallel()

	const fps = 60
	h := newParkHarness(t, fps)
	h.waitParked()

	stop := make(chan struct{})
	go func() {
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			case <-time.After(50 * time.Millisecond):
				h.setView(func(v *View) { v.Content = fmt.Sprintf("frame %d", i) })
			}
		}
	}()

	time.Sleep(200 * time.Millisecond) // let it get going
	n := h.r.count()
	time.Sleep(time.Second)
	ticks := h.r.count() - n
	close(stop)

	// The view changes every third frame, which is far too often to park.
	// Allow for scheduling noise, but parking would cut this to about 40.
	if ticks < fps*8/10 {
		t.Errorf("ticked %d times in a second while active, want about %d", ticks, fps)
	}
}

// A woken renderer draws on the same frame boundaries as a ticker that never
// stopped, so the first frame after idle is no later than it used to be.
func TestRendererWakesOnFrameBoundary(t *testing.T) {
	t.Parallel()

	const fps = 10
	h := newParkHarness(t, fps)
	h.waitParked()

	n := h.r.count()
	last := h.r.after(n - 1)[0]

	// Change the view half way between two frame boundaries. A ticker that
	// restarted from the change would draw half a frame off the boundary.
	time.Sleep(h.frame/2 - time.Since(last)%h.frame + h.frame)
	mark := h.out.Len()
	sent := time.Now()
	h.setView(func(v *View) { v.Content = "second" })
	h.waitOutput(mark, 5*time.Second, "second")

	woke := h.r.after(n)
	if len(woke) == 0 {
		t.Fatal("no tick after the wake")
	}
	first := woke[0]
	if d := first.Sub(sent); d > h.frame+50*time.Millisecond {
		t.Errorf("first frame after idle drawn %v after the change, want at most one frame (%v)", d, h.frame)
	}

	t.Logf("sent %v after last tick, first frame %v after last tick", sent.Sub(last), first.Sub(last))
	// Distance from the nearest frame boundary of the old ticker.
	off := first.Sub(last) % h.frame
	if off > h.frame/2 {
		off = h.frame - off
	}
	if off > 25*time.Millisecond {
		t.Errorf("first frame after idle is %v off the frame boundary", off)
	}
}

// Every way of giving the renderer work must wake it.
func TestRendererWakes(t *testing.T) {
	t.Parallel()

	red := color.RGBA{R: 0xff, A: 0xff}
	view := func(fn func(*View)) func(h *parkHarness) {
		return func(h *parkHarness) { h.setView(fn) }
	}
	send := func(msgs ...Msg) func(h *parkHarness) {
		return func(h *parkHarness) {
			for _, msg := range msgs {
				h.p.Send(msg)
			}
		}
	}
	cmd := func(c Cmd) func(h *parkHarness) {
		return func(h *parkHarness) { h.p.Send(runCmdMsg{c}) }
	}

	tests := []struct {
		name string
		do   func(h *parkHarness)
		want []string
	}{
		// Changes to the view.
		{"content", view(func(v *View) { v.Content = "second" }), []string{"second"}},
		{"cursor", view(func(v *View) { v.Cursor = NewCursor(2, 0) }), []string{ansi.SetModeTextCursorEnable}},
		{"cursor shape", view(func(v *View) {
			v.Cursor = NewCursor(0, 0)
			v.Cursor.Shape = CursorBar
		}), []string{ansi.SetCursorStyle(encodeCursorStyle(CursorBar, true))}},
		{"cursor color", view(func(v *View) {
			v.Cursor = NewCursor(0, 0)
			v.Cursor.Color = red
		}), []string{ansi.SetCursorColor("#ff0000")}},
		{"window title", view(func(v *View) { v.WindowTitle = "title" }), []string{ansi.SetWindowTitle("title")}},
		{"mouse cell motion", view(func(v *View) { v.MouseMode = MouseModeCellMotion }), []string{ansi.SetModeMouseButtonEvent}},
		{"mouse all motion", view(func(v *View) { v.MouseMode = MouseModeAllMotion }), []string{ansi.SetModeMouseAnyEvent}},
		{"report focus", view(func(v *View) { v.ReportFocus = true }), []string{ansi.SetModeFocusEvent}},
		{"bracketed paste", view(func(v *View) { v.DisableBracketedPasteMode = true }), []string{ansi.ResetModeBracketedPaste}},
		{"keyboard enhancements", view(func(v *View) { v.KeyboardEnhancements.ReportEventTypes = true }), []string{
			ansi.KittyKeyboard(keyboardEnhancementsFlags(KeyboardEnhancements{ReportEventTypes: true}), 1),
		}},
		{"alt screen", view(func(v *View) { v.AltScreen = true }), []string{ansi.SetModeAltScreenSaveCursor, "first"}},
		{"foreground color", view(func(v *View) { v.ForegroundColor = red }), []string{ansi.SetForegroundColor("#ff0000")}},
		{"background color", view(func(v *View) { v.BackgroundColor = red }), []string{ansi.SetBackgroundColor("#ff0000")}},
		{"progress bar", view(func(v *View) { v.ProgressBar = NewProgressBar(ProgressBarDefault, 50) }), []string{ansi.SetProgressBar(50)}},
		{"mouse handler", func(h *parkHarness) {
			// The renderer only picks up a new handler with a frame that
			// changed, so change the content along with it.
			mark := h.out.Len()
			h.setView(func(v *View) {
				v.Content = "........"
				v.OnMouse = func(MouseMsg) Cmd {
					return func() Msg { return setViewMsg(func(v *View) { v.Content = "CLICKED" }) }
				}
			})
			h.waitOutput(mark, 2*time.Second, "........")
			h.waitParked()
			h.p.Send(MouseClickMsg{X: 1, Y: 0})
		}, []string{"CLICKED"}},

		// Lines printed above the program.
		{"Program.Println", func(h *parkHarness) { h.p.Println("printed") }, []string{"printed"}},
		{"Program.Printf", func(h *parkHarness) { h.p.Printf("printed %d", 2) }, []string{"printed 2"}},
		{"Println", cmd(Println("printed")), []string{"printed"}},
		{"Printf", cmd(Printf("printed %d", 3)), []string{"printed 3"}},

		// Screen changes that redraw the whole frame.
		{"window size", send(WindowSizeMsg{Width: 100, Height: 30}), []string{"first"}},
		{"clear screen", send(ClearScreen()), []string{"first"}},

		// Sequences queued for the output.
		{"raw", cmd(Raw("\x1b]2;raw\x07")), []string{"\x1b]2;raw\x07"}},
		{"set clipboard", cmd(SetClipboard("clip")), []string{ansi.SetSystemClipboard("clip")}},
		{"set primary clipboard", cmd(SetPrimaryClipboard("clip")), []string{ansi.SetPrimaryClipboard("clip")}},
		{"read clipboard", cmd(ReadClipboard), []string{ansi.RequestSystemClipboard}},
		{"read primary clipboard", cmd(ReadPrimaryClipboard), []string{ansi.RequestPrimaryClipboard}},
		{"background color query", cmd(RequestBackgroundColor), []string{ansi.RequestBackgroundColor}},
		{"foreground color query", cmd(RequestForegroundColor), []string{ansi.RequestForegroundColor}},
		{"cursor color query", cmd(RequestCursorColor), []string{ansi.RequestCursorColor}},
		{"cursor position query", cmd(RequestCursorPosition), []string{ansi.RequestCursorPositionReport}},
		{"terminal version query", cmd(RequestTerminalVersion), []string{ansi.RequestNameVersion}},
		{"capability query", cmd(RequestCapability("RGB")), []string{ansi.RequestTermcap("RGB")}},

		// Terminal state the renderer applies to the next frame.
		{"synchronized output", send(
			ModeReportMsg{Mode: ansi.ModeSynchronizedOutput, Value: ansi.ModeReset},
			setViewMsg(func(v *View) { v.Content = "second" }),
		), []string{ansi.SetModeSynchronizedOutput, "second"}},
		{"unicode core", send(
			ModeReportMsg{Mode: ansi.ModeUnicodeCore, Value: ansi.ModeSet},
			setViewMsg(func(v *View) { v.Content = "second" }),
		), []string{ansi.SetModeUnicodeCore, "second"}},
		{"color profile", send(
			ColorProfileMsg{colorprofile.ANSI},
			setViewMsg(func(v *View) { v.Content = "second" }),
		), []string{"second"}},

		// Sequences queued from outside the event loop.
		{"execute", func(h *parkHarness) { h.p.execute("\x1b]2;direct\x07") }, []string{"\x1b]2;direct\x07"}},

		// Handing the terminal to another process and taking it back.
		{"exec", cmd(Exec(nopExecCommand{}, nil)), []string{ansi.SetModeBracketedPaste, "first"}},
		{"release and restore", func(h *parkHarness) {
			if err := h.p.ReleaseTerminal(); err != nil {
				h.t.Fatal(err)
			}
			if err := h.p.RestoreTerminal(); err != nil {
				h.t.Fatal(err)
			}
		}, []string{ansi.SetModeBracketedPaste, "first"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newParkHarness(t, 60)
			h.waitOutput(0, 5*time.Second, "first")
			h.waitParked()

			mark := h.out.Len()
			tc.do(h)
			took := h.waitOutput(mark, 2*time.Second, tc.want...)
			t.Logf("drawn %v after waking", took)

			// And it parks again afterwards.
			h.waitParked()

			// A second change is drawn too.
			mark = h.out.Len()
			h.setView(func(v *View) { v.Content = "AGAIN" })
			h.waitOutput(mark, 2*time.Second, "AGAIN")
		})
	}
}

func TestRendererShutdownWhileParked(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		do      func(h *parkHarness)
		wantErr error
		want    string // final frame, if drawn
	}{
		{"quit", func(h *parkHarness) {
			h.setView(func(v *View) { v.Content = "last" })
			h.p.Quit()
		}, nil, "last"},
		{"quit cmd", func(h *parkHarness) { h.p.Send(runCmdMsg{Quit}) }, nil, ""},
		{"kill", func(h *parkHarness) { h.p.Kill() }, ErrProgramKilled, ""},
		{"interrupt", func(h *parkHarness) { h.p.Send(Interrupt()) }, ErrInterrupted, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newParkHarness(t, 60)
			h.waitParked()

			mark := h.out.Len()
			tc.do(h)
			err := h.waitDone()
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Run returned %v, want %v", err, tc.wantErr)
			}
			if tc.want != "" && !strings.Contains(h.out.since(mark), tc.want) {
				t.Errorf("final frame %q not drawn", tc.want)
			}
		})
	}

	t.Run("context", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		h := newParkHarness(t, 60, WithContext(ctx))
		h.waitParked()
		cancel()
		if err := h.waitDone(); !errors.Is(err, ErrProgramKilled) {
			t.Errorf("Run returned %v, want %v", err, ErrProgramKilled)
		}
	})
}

// rendererGoroutines counts the running renderer goroutines.
func rendererGoroutines() int {
	buf := make([]byte, 1<<20)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return strings.Count(string(buf[:n]), "(*Program).startRenderer.func")
		}
		buf = make([]byte, 2*len(buf))
	}
}

// Stopping and restarting a parked renderer must not leave goroutines behind.
// Not parallel: it counts goroutines across the whole process.
func TestRendererParkedNoGoroutineLeak(t *testing.T) {
	before := rendererGoroutines()

	for _, stop := range []string{"quit", "kill"} {
		h := newParkHarness(t, 60)
		h.waitParked()

		// Restart the renderer while parked, both ways.
		h.p.Send(runCmdMsg{Exec(nopExecCommand{}, nil)})
		h.waitParked()
		if err := h.p.ReleaseTerminal(); err != nil {
			t.Fatal(err)
		}
		if err := h.p.RestoreTerminal(); err != nil {
			t.Fatal(err)
		}
		h.waitParked()

		if got := rendererGoroutines() - before; got != 1 {
			t.Errorf("%s: %d renderer goroutines while running, want 1", stop, got)
		}

		if stop == "quit" {
			h.p.Quit()
		} else {
			h.p.Kill()
		}
		_ = h.waitDone()
	}

	deadline := time.Now().Add(2 * time.Second)
	for rendererGoroutines() != before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := rendererGoroutines() - before; got != 0 {
		t.Errorf("%d renderer goroutines left after the programs exited", got)
	}
}

// Wakes race with parking from many goroutines at once. The last change must
// always be drawn. Run with -race.
func TestRendererParkStress(t *testing.T) {
	t.Parallel()

	h := newParkHarness(t, maxFPS)
	h.waitOutput(0, 5*time.Second, "first")

	// Bursts of every kind of wake from several goroutines, with gaps long
	// enough for the renderer to park in between.
	for round := range 3 {
		var wg sync.WaitGroup
		stop := make(chan struct{})
		worker := func(every time.Duration, fn func(i int)) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; ; i++ {
					select {
					case <-stop:
						return
					case <-time.After(every):
						fn(i)
					}
				}
			}()
		}
		for w := range 4 {
			worker(time.Duration(w+1)*3*time.Millisecond, func(i int) {
				h.setView(func(v *View) { v.Content = fmt.Sprintf("r%d w%d %d", round, w, i) })
			})
		}
		worker(7*time.Millisecond, func(i int) { h.p.Println("line", i) })
		worker(11*time.Millisecond, func(i int) { h.p.Send(WindowSizeMsg{Width: 80 + i%5, Height: 24}) })
		worker(13*time.Millisecond, func(i int) { h.p.Send(runCmdMsg{Raw(fmt.Sprintf("\x1b]2;%d\x07", i))}) })
		worker(50*time.Millisecond, func(int) { h.p.Send(runCmdMsg{Exec(nopExecCommand{}, nil)}) })

		time.Sleep(300 * time.Millisecond)
		close(stop)
		wg.Wait()

		end := fmt.Sprintf("end of round %d", round)
		h.setView(func(v *View) { v.Content = end })
		h.waitDrawn(end, 2*time.Second)
		h.waitParked()
	}

	h.waitParked()
	h.setView(func(v *View) { v.Content = "final" })
	h.waitDrawn("final", 2*time.Second)

	h.p.Quit()
	if err := h.waitDone(); err != nil {
		t.Fatal(err)
	}
}
