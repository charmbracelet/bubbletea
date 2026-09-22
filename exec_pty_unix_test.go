//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix zos

package tea

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestExecBridgedClipboardSet(t *testing.T) {
	t.Parallel()

	backend := &stubClipboardBackend{}
	var out bytes.Buffer
	p := NewProgram(&testModel{},
		WithWindowSize(80, 24),
		WithInput(strings.NewReader("")),
		WithOutput(&out),
		WithClipboardBackend(backend),
	)

	// The command copies "hello" with an OSC52 sequence surrounded by plain
	// output. The sequence must be intercepted, not forwarded.
	cmd := exec.Command("sh", "-c", `printf 'before\033]52;c;aGVsbG8=\a after'`)
	if err := p.execBridged(cmd); err != nil {
		t.Fatalf("execBridged: %v", err)
	}

	if got := out.String(); got != "before after" {
		t.Errorf("output = %q, want %q", got, "before after")
	}
	calls := backend.setCallsFor()
	if len(calls) != 1 || calls[0].content != "hello" {
		t.Errorf("set calls = %v, want one call with %q", calls, "hello")
	}
}

func TestExecBridgedClipboardRead(t *testing.T) {
	t.Parallel()

	backend := &stubClipboardBackend{getValues: map[byte]string{uv.SystemClipboard: "hello"}}
	var out bytes.Buffer

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	p := NewProgram(&testModel{},
		WithContext(ctx),
		WithWindowSize(80, 24),
		WithInput(strings.NewReader("")),
		WithOutput(&out),
		WithClipboardBackend(backend),
	)

	// The command asks for the clipboard, then reads exactly as many bytes
	// as the reply takes. Raw mode is needed so the read returns without a
	// line delimiter, and echo is disabled so the reply is not echoed back
	// as command output. If the reply is never injected, the command blocks
	// until the context times out.
	cmd := exec.Command("sh", "-c", `stty raw -echo 2>/dev/null; printf '\033]52;c;?\a'; head -c 16 >/dev/null`)

	if err := p.execBridged(cmd); err != nil {
		t.Fatalf("execBridged: %v", err)
	}

	if got := backend.getCallsFor(); len(got) != 1 || got[0] != uv.SystemClipboard {
		t.Errorf("get calls = %v, want one call for %q", got, uv.SystemClipboard)
	}
	if got := out.String(); got != "" {
		t.Errorf("output = %q, want empty", got)
	}
}

func TestExecBridgedPlainOutput(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	p := NewProgram(&testModel{},
		WithWindowSize(80, 24),
		WithInput(strings.NewReader("")),
		WithOutput(&out),
		WithClipboardBackend(&stubClipboardBackend{}),
	)

	cmd := exec.Command("sh", "-c", `printf 'plain output'`)
	if err := p.execBridged(cmd); err != nil {
		t.Fatalf("execBridged: %v", err)
	}
	if got := out.String(); got != "plain output" {
		t.Errorf("output = %q, want %q", got, "plain output")
	}
}
