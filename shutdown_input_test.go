package tea

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"testing/synctest"
	"time"
)

type shutdownInputReader struct {
	cancelable bool
	cancels    int
	closes     int
}

func (*shutdownInputReader) Read([]byte) (int, error) { return 0, io.EOF }
func (r *shutdownInputReader) Cancel() bool           { r.cancels++; return r.cancelable }
func (r *shutdownInputReader) Close() error           { r.closes++; return nil }

func TestShutdownWaitsForCancelableInput(t *testing.T) {
	for _, kill := range []bool{false, true} {
		name := "quit"
		if kill {
			name = "kill"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				p := NewProgram(&testModel{}, WithOutput(io.Discard))
				r := &shutdownInputReader{cancelable: true}
				p.cancelReader = r
				p.readLoopDone = make(chan struct{})
				stopped := make(chan struct{})
				go func() { p.shutdown(kill); close(stopped) }()
				synctest.Wait()
				if r.cancels != 1 || r.closes != 0 {
					t.Errorf("before read loop completion: cancel=%d close=%d", r.cancels, r.closes)
				}
				select {
				case <-stopped:
					t.Error("shutdown returned before the input loop finished")
				default:
				}
				close(p.readLoopDone)
				synctest.Wait()
				<-stopped
				if r.closes != 1 {
					t.Errorf("close calls after input completion: got %d, want 1", r.closes)
				}
				p.shutdown(kill)
				if r.cancels != 1 || r.closes != 1 {
					t.Errorf("repeated shutdown: cancel=%d close=%d", r.cancels, r.closes)
				}
			})
		})
	}
}

func TestShutdownDefersCloseAfterInputTimeout(t *testing.T) {
	for _, kill := range []bool{false, true} {
		name := "quit"
		if kill {
			name = "kill"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				p := NewProgram(&testModel{}, WithOutput(io.Discard))
				r := &shutdownInputReader{cancelable: true}
				p.cancelReader = r
				done := make(chan struct{})
				p.readLoopDone = done
				stopped := make(chan struct{})
				go func() { p.shutdown(kill); close(stopped) }()
				synctest.Wait()
				time.Sleep(time.Second)
				synctest.Wait()
				select {
				case <-stopped:
				default:
					t.Error("shutdown did not retain its bounded wait")
				}
				if r.closes != 0 {
					t.Errorf("reader closed while its input loop was still active: %d", r.closes)
				}
				// Cleanup must retain this reader, not follow a later field assignment.
				replacement := &shutdownInputReader{}
				p.cancelReader = replacement
				p.readLoopDone = make(chan struct{})
				close(done)
				synctest.Wait()
				if r.closes != 1 || replacement.closes != 0 {
					t.Errorf("close calls: original=%d replacement=%d", r.closes, replacement.closes)
				}
				<-stopped
			})
		})
	}
}

func TestShutdownInputFallbacks(t *testing.T) {
	for _, tc := range []struct {
		name       string
		reader     bool
		cancelable bool
		loop       bool
		finished   bool
	}{
		{name: "no reader"},
		{name: "uncancelable", reader: true, loop: true},
		{name: "no input loop", reader: true, cancelable: true},
		{name: "finished loop", reader: true, cancelable: true, loop: true, finished: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				p := NewProgram(&testModel{}, WithOutput(io.Discard))
				r := &shutdownInputReader{cancelable: tc.cancelable}
				if tc.reader {
					p.cancelReader = r
				}
				if tc.loop {
					p.readLoopDone = make(chan struct{})
					if tc.finished {
						close(p.readLoopDone)
					}
				}
				start := time.Now()
				p.shutdown(true)
				if time.Since(start) != 0 {
					t.Error("input fallback unexpectedly waited")
				}
				if tc.reader && (r.cancels != 1 || r.closes != 1) {
					t.Errorf("fallback cleanup: cancel=%d close=%d", r.cancels, r.closes)
				}
			})
		})
	}
}

type inputStartedMsg struct{}
type inputShutdownModel struct{ ready chan struct{} }

func (m inputShutdownModel) Init() Cmd { return func() Msg { return inputStartedMsg{} } }
func (m inputShutdownModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(inputStartedMsg); ok {
		close(m.ready)
	}
	return m, nil
}
func (inputShutdownModel) View() View { return NewView("ready") }

func TestProgramCanceledInput(t *testing.T) {
	for _, kill := range []bool{false, true} {
		name := "context"
		if kill {
			name = "kill"
		}
		t.Run(name, func(t *testing.T) {
			// A real file exercises the native cancelreader rather than its no-op fallback.
			reader, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()
			defer writer.Close()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			ready := make(chan struct{})
			p := NewProgram(inputShutdownModel{ready}, WithContext(ctx), WithInput(reader), WithOutput(io.Discard), WithWindowSize(80, 24), WithoutSignalHandler())
			result := make(chan error, 1)
			go func() { _, err := p.Run(); result <- err }()
			select {
			case <-ready:
			case <-time.After(5 * time.Second):
				t.Fatal("event loop did not start")
			}
			if kill {
				p.Kill()
			} else {
				cancel()
			}
			select {
			case err = <-result:
			case <-time.After(5 * time.Second):
				t.Fatal("program did not stop")
			}
			if !errors.Is(err, ErrProgramKilled) {
				t.Fatalf("unexpected error: %v", err)
			}
			if !kill && !errors.Is(err, context.Canceled) {
				t.Fatalf("missing external cancellation: %v", err)
			}
			if _, err := reader.Stat(); err != nil {
				t.Fatalf("borrowed input was closed: %v", err)
			}
		})
	}
}
