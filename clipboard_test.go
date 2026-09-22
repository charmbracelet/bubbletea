package tea

import (
	"bytes"
	"context"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

// stubClipboardCall records a single clipboard operation.
type stubClipboardCall struct {
	selection byte
	content   string
}

// stubClipboardBackend is an in-memory [ClipboardBackend] for tests.
type stubClipboardBackend struct {
	mu        sync.Mutex
	setCalls  []stubClipboardCall
	getCalls  []byte
	getValues map[byte]string
	setErr    error
	getErr    error
}

func (b *stubClipboardBackend) Set(selection ClipboardSelection, content string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.setCalls = append(b.setCalls, stubClipboardCall{selection: selection, content: content})
	return b.setErr
}

func (b *stubClipboardBackend) Get(selection ClipboardSelection) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.getCalls = append(b.getCalls, selection)
	if b.getErr != nil {
		return "", b.getErr
	}
	return b.getValues[selection], nil
}

func (b *stubClipboardBackend) setCallsFor() []stubClipboardCall {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]stubClipboardCall(nil), b.setCalls...)
}

func (b *stubClipboardBackend) getCallsFor() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]byte(nil), b.getCalls...)
}

// withTestTimeout returns an option that bounds the runtime of a program under
// test, so a bug that prevents it from exiting fails the test instead of
// hanging it.
func withTestTimeout(t *testing.T) ProgramOption {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)
	return WithContext(ctx)
}

// clipboardTestModel quits as soon as it receives a clipboard message, so
// tests don't need to poll for it.
type clipboardTestModel struct {
	received chan ClipboardMsg
}

func (m *clipboardTestModel) Init() Cmd { return nil }

func (m *clipboardTestModel) Update(msg Msg) (Model, Cmd) {
	if cm, ok := msg.(ClipboardMsg); ok {
		m.received <- cm
		return m, Quit
	}
	return m, nil
}

func (m *clipboardTestModel) View() View { return NewView("") }

func TestWithClipboardBackendIsUsed(t *testing.T) {
	t.Parallel()

	backend := &stubClipboardBackend{getValues: map[byte]string{uv.SystemClipboard: "pasted"}}
	model := &clipboardTestModel{received: make(chan ClipboardMsg, 1)}

	var out bytes.Buffer
	p := NewProgram(model,
		WithWindowSize(80, 24),
		WithEnvironment([]string{"TERM=xterm-256color"}),
		WithInput(&bytes.Buffer{}),
		WithOutput(&out),
		WithClipboardBackend(backend),
		withTestTimeout(t),
	)

	go p.Send(sequenceMsg([]Cmd{SetClipboard("copied"), ReadClipboard}))

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if got := out.String(); strings.Contains(got, "\x1b]52;") {
		t.Errorf("output contains an OSC52 sequence: %q", got)
	}

	// The set may still be in flight when the read completes, so wait for it.
	deadline := time.Now().Add(time.Second)
	for len(backend.setCallsFor()) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	calls := backend.setCallsFor()
	if len(calls) != 1 || calls[0].content != "copied" {
		t.Errorf("set calls = %v, want one call with %q", calls, "copied")
	}

	select {
	case cm := <-model.received:
		if cm.Content != "pasted" {
			t.Errorf("ClipboardMsg content = %q, want %q", cm.Content, "pasted")
		}
		if cm.Clipboard() != uv.SystemClipboard {
			t.Errorf("ClipboardMsg selection = %q, want %q", cm.Clipboard(), uv.SystemClipboard)
		}
	default:
		t.Error("no ClipboardMsg received")
	}
}

func TestClipboardWithoutBackendUsesOSC52(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	p := NewProgram(&testModel{},
		WithWindowSize(80, 24),
		WithOutput(&out),
	)

	p.handleSetClipboard(uv.SystemClipboard, "hello")
	p.handleReadClipboard(uv.SystemClipboard)
	p.handleSetClipboard(uv.PrimaryClipboard, "primary")
	p.handleReadClipboard(uv.PrimaryClipboard)

	if err := p.flush(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	// base64("hello") is "aGVsbG8=", base64("primary") is "cHJpbWFyeQ==".
	for _, want := range []string{"\x1b]52;c;aGVsbG8=", "\x1b]52;c;?", "\x1b]52;p;cHJpbWFyeQ==", "\x1b]52;p;?"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q does not contain %q", got, want)
		}
	}
}

func TestClipboardBackendErrorsDoNotHangReads(t *testing.T) {
	t.Parallel()

	backend := &stubClipboardBackend{getErr: ErrClipboardUnavailable}
	model := &clipboardTestModel{received: make(chan ClipboardMsg, 1)}

	var out bytes.Buffer
	p := NewProgram(model,
		WithWindowSize(80, 24),
		WithEnvironment([]string{"TERM=xterm-256color"}),
		WithInput(&bytes.Buffer{}),
		WithOutput(&out),
		WithClipboardBackend(backend),
		withTestTimeout(t),
	)

	go p.Send(sequenceMsg([]Cmd{ReadClipboard}))

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}

	select {
	case cm := <-model.received:
		if cm.Content != "" {
			t.Errorf("ClipboardMsg content = %q, want empty", cm.Content)
		}
	default:
		t.Error("no ClipboardMsg received")
	}
}

func TestClipboardFallbackDetection(t *testing.T) {
	t.Parallel()

	t.Run("apple terminal selects a backend on darwin", func(t *testing.T) {
		t.Parallel()
		p := NewProgram(&testModel{}, WithEnvironment([]string{"TERM_PROGRAM=Apple_Terminal", "TERM=xterm-256color"}))
		if runtime.GOOS == "darwin" {
			if p.clipboard == nil {
				t.Error("expected a clipboard backend on darwin")
			}
		} else if p.clipboard != nil {
			t.Error("expected no clipboard backend without a terminal clipboard tool")
		}
	})

	t.Run("supported terminal does not select a backend", func(t *testing.T) {
		t.Parallel()
		p := NewProgram(&testModel{}, WithEnvironment([]string{"TERM_PROGRAM=iTerm.app", "TERM=xterm-256color"}))
		if p.clipboard != nil {
			t.Error("expected no clipboard backend on a terminal with OSC52 support")
		}
	})

	t.Run("fallback can be disabled", func(t *testing.T) {
		t.Parallel()
		p := NewProgram(&testModel{},
			WithEnvironment([]string{"TERM_PROGRAM=Apple_Terminal", "TERM=xterm-256color"}),
			WithoutClipboardFallback(),
		)
		if p.clipboard != nil {
			t.Error("expected no clipboard backend when the fallback is disabled")
		}
	})
}
