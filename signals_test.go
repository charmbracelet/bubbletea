//go:build !windows

package tea

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

// TestHandleSignalsReturnsAfterContextCancelDuringPendingSend is a
// regression test for the signal-goroutine deadlock on shutdown: a signal
// arriving while nobody reads p.msgs (the event loop already stopped) must
// not block the handler forever once the program context is canceled.
func TestHandleSignalsReturnsAfterContextCancelDuringPendingSend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register our own handler first so the SIGINT can never take down the
	// test process before handleSignals registers its own.
	ours := make(chan os.Signal, 1)
	signal.Notify(ours, syscall.SIGINT)
	defer signal.Stop(ours)

	p := &Program{
		ctx:  ctx,
		msgs: make(chan Msg), // unbuffered, no reader: simulates a stopped event loop
	}

	done := p.handleSignals()
	time.Sleep(100 * time.Millisecond) // let the handler register its signal.Notify

	if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}
	select {
	case <-ours:
	case <-time.After(5 * time.Second):
		t.Fatal("SIGINT was not delivered")
	}
	// Let the handler consume its copy of the signal and enter the send on
	// p.msgs, where it would block forever before the fix.
	time.Sleep(100 * time.Millisecond)

	// Canceling the context (what shutdown does) must unblock the handler.
	cancel()

	select {
	case <-done:
		// The handler returned; shutdown can proceed.
	case <-time.After(5 * time.Second):
		t.Fatal("handleSignals did not return after context cancel: signal goroutine is deadlocked")
	}
}

// TestHandleSignalsCatchesSecondSignalDuringShutdown is a regression test for
// https://github.com/charmbracelet/bubbletea/issues/1849: after a quit
// message is forwarded, the signal subscription must stay alive until the
// program finishes, so a second SIGINT during shutdown is caught instead of
// taking the runtime's default path and killing the process.
func TestHandleSignalsCatchesSecondSignalDuringShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register our own handler first so a SIGINT can never take down the
	// test process, whether or not handleSignals is catching it.
	ours := make(chan os.Signal, 2)
	signal.Notify(ours, syscall.SIGINT)
	defer signal.Stop(ours)

	finished := make(chan struct{})
	defer close(finished)

	p := &Program{
		ctx:      ctx,
		msgs:     make(chan Msg),
		finished: finished,
	}

	done := p.handleSignals()
	time.Sleep(100 * time.Millisecond) // let the handler register its signal.Notify

	if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}
	<-ours

	// The first signal is forwarded as an interrupt message.
	select {
	case msg := <-p.msgs:
		if _, ok := msg.(InterruptMsg); !ok {
			t.Fatalf("expected InterruptMsg, got %T", msg)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("first SIGINT was not forwarded")
	}

	// Context cancellation (the start of shutdown) ends the handler.
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handleSignals did not finish after context cancel")
	}

	// Release our own watcher so that whether SIGINT is still caught depends
	// solely on the program's subscription. A second SIGINT during the
	// shutdown window must still be caught (and dropped); if the program had
	// already released its subscription it would take the runtime's default
	// path and terminate this test process — so surviving is the assertion.
	signal.Stop(ours)
	if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("failed to send second SIGINT: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
}
