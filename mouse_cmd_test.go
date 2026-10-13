package tea

// TestMouseCmdBlocksEventLoop demonstrates a real behavioral quirk in the
// eventLoop's MouseMsg branch (tea.go):
//
//	if cmd := p.renderer.onMouse(msg); cmd != nil {
//	    go p.Send(cmd())
//	}
//
// Because Go evaluates the arguments of `go f(args)` in the *current*
// goroutine before spawning the new one, the user-supplied Cmd `cmd()` here
// runs synchronously *on the event loop goroutine*. Any blocking work inside
// that Cmd therefore stalls the whole event loop, delaying every other
// message — including Quit — until the Cmd returns.
//
// Every other user-Cmd execution path (handleCommands, execBatchMsg,
// execSequenceMsg) instead runs `go func() { msg := cmd(); ... }()`, which
// keeps the blocking Cmd off the event loop.
//
// This test runs the scenario in a subprocess: the helper's OnMouse Cmd sleeps
// for ~300ms and we measure whether Quit is delayed by that amount. With the
// current code the event loop is blocked and the subprocess exits non-zero
// (reproducing the problem); once the branch is fixed to execute the Cmd
// asynchronously, Quit is handled immediately and the subprocess exits 0.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"
	"testing"
	"time"
)

// mouseBlockDuration is how long the OnMouse Cmd sleeps, i.e. how long the
// event loop would be stalled while the Cmd runs synchronously.
const mouseBlockDuration = 300 * time.Millisecond

// mouseBlockDoneMsg is returned by the blocking Cmd after it finishes.
type mouseBlockDoneMsg struct{}

// blockingCmd sleeps for mouseBlockDuration before returning.
func blockingCmd() Msg {
	time.Sleep(mouseBlockDuration)
	return mouseBlockDoneMsg{}
}

// blockingRenderer is a controllable renderer whose onMouse returns the
// blocking Cmd. It stands in for a real renderer whose lastView was produced
// by user code (View.OnMouse), letting us drive the eventLoop's MouseMsg
// branch deterministically without depending on renderer flush scheduling.
type blockingRenderer struct {
	nilRenderer
}

func (r *blockingRenderer) onMouse(MouseMsg) Cmd {
	return blockingCmd
}

// blockingModel is a minimal model; rendering it signals that the event loop
// is up and running.
type blockingModel struct {
	rendered atomic.Bool
}

func (m *blockingModel) Init() Cmd { return nil }

func (m *blockingModel) Update(Msg) (Model, Cmd) { return m, nil }

func (m *blockingModel) View() View {
	m.rendered.Store(true)
	return NewView("hello")
}

// runBlockingHelper is the subprocess body. It fires a MouseClickMsg (whose
// Cmd sleeps), immediately followed by Quit, and exits 0 if Quit was handled
// promptly, 1 if the event loop was blocked by the mouse Cmd.
func runBlockingHelper() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var out bytes.Buffer
	m := &blockingModel{}
	p := NewProgram(m,
		WithContext(ctx),
		WithInput(&bytes.Buffer{}),
		WithOutput(&out),
	)
	// Inject a renderer whose onMouse returns the blocking Cmd. (Internal
	// test: we can reach Program's private fields.)
	p.renderer = &blockingRenderer{}

	start := time.Now()
	go func() {
		for !m.rendered.Load() {
			time.Sleep(time.Millisecond)
		}
		// Fire the mouse event, then immediately ask to quit.
		p.Send(MouseClickMsg{})
		p.Send(QuitMsg{})
	}()
	_, _ = p.Run()
	elapsed := time.Since(start)
	fmt.Fprintf(os.Stderr, "HELPER DIAG: elapsed since mouse event = %v\n", elapsed)

	// If the mouse Cmd blocked the event loop, Quit was delayed by roughly
	// mouseBlockDuration (300ms). Give a comfortable margin: under 250ms the
	// event loop was not stalled.
	if elapsed > 250*time.Millisecond {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestMouseCmdBlocksEventLoop(t *testing.T) {
	if os.Getenv("BT_MOUSE_HELPER") == "1" {
		runBlockingHelper()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMouseCmdBlocksEventLoop$")
	cmd.Env = append(os.Environ(), "BT_MOUSE_HELPER=1")
	out, err := cmd.CombinedOutput()

	if err == nil {
		t.Logf("OK: Quit was not delayed by the mouse Cmd (event loop not blocked). Subprocess output:\n%s", out)
		return
	}

	if ee, ok := err.(*exec.ExitError); ok && !ee.Success() {
		t.Fatalf(
			"BUG REPRODUCED: the Cmd returned by View.OnMouse blocked the event loop, delaying Quit by ~%v.\n"+
				"Expected the Cmd to run without stalling the event loop (as handleCommands/execBatchMsg do).\n"+
				"Subprocess output:\n%s",
			mouseBlockDuration, out,
		)
	}
	t.Fatalf("unexpected error running helper subprocess: %v\n%s", err, out)
}
