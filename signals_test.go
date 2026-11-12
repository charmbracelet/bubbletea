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
