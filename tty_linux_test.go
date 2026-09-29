//go:build linux
// +build linux

package tea

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type readyMsg struct{}

type readyModel struct{ ready chan struct{} }

func (m readyModel) Init() Cmd { return func() Msg { return readyMsg{} } }

func (m readyModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(readyMsg); ok {
		close(m.ready)
	}
	return m, nil
}

func (m readyModel) View() View { return NewView("") }

func countEpollFDs(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Skipf("cannot inspect file descriptors: %v", err)
	}
	var n int
	for _, e := range entries {
		target, err := os.Readlink(filepath.Join("/proc/self/fd", e.Name()))
		if err == nil && target == "anon_inode:[eventpoll]" {
			n++
		}
	}
	return n
}

// Every ReleaseTerminal/RestoreTerminal cycle (Exec, suspend) replaces the
// input cancel reader. The replaced reader must be closed, or its epoll file
// descriptor leaks.
//
// Not parallel: it counts process-wide file descriptors.
func TestReleaseRestoreTerminalClosesCancelReader(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close() //nolint:errcheck
	defer w.Close() //nolint:errcheck

	m := readyModel{ready: make(chan struct{})}
	var out bytes.Buffer
	p := NewProgram(m, WithInput(r), WithOutput(&out))
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = p.Run()
	}()
	select {
	case <-m.ready:
	case <-time.After(5 * time.Second):
		t.Fatal("program did not start")
	}

	before := countEpollFDs(t)
	const cycles = 20
	for i := range cycles {
		if err := p.ReleaseTerminal(); err != nil {
			t.Fatalf("release %d: %v", i, err)
		}
		if err := p.RestoreTerminal(); err != nil {
			t.Fatalf("restore %d: %v", i, err)
		}
	}
	after := countEpollFDs(t)

	p.Quit()
	<-done

	if after > before {
		t.Fatalf("epoll file descriptors grew by %d over %d release/restore cycles (before=%d, after=%d)",
			after-before, cycles, before, after)
	}
}
