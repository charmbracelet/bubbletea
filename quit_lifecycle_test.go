package tea

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

type quitLifecycleModel struct {
	init     func()
	onUpdate func(Msg) Cmd
}

func (m quitLifecycleModel) Init() Cmd {
	if m.init != nil {
		m.init()
	}
	return nil
}

func (m quitLifecycleModel) Update(msg Msg) (Model, Cmd) {
	if m.onUpdate != nil {
		return m, m.onUpdate(msg)
	}
	return m, nil
}

func (quitLifecycleModel) View() View { return NewView("running") }

func awaitQuitLifecycle(t *testing.T, done <-chan struct{}, operation string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %s", operation)
	}
}

func TestProgramQuitUnstarted(t *testing.T) {
	t.Parallel()
	for _, model := range []Model{nil, quitLifecycleModel{}} {
		t.Run("model", func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			p := NewProgram(model, WithContext(ctx))
			done := make(chan struct{})
			go func() {
				p.Quit()
				p.Quit()
				close(done)
			}()
			defer func() {
				cancel()
				awaitQuitLifecycle(t, done, "unstarted Quit cleanup")
			}()
			awaitQuitLifecycle(t, done, "unstarted Quit")
			if err := p.ctx.Err(); err != nil {
				t.Fatalf("Quit canceled an unstarted program: %v", err)
			}
		})
	}
}

func TestProgramQuitBeforeRunPreservesMessages(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	var received bool
	m := quitLifecycleModel{onUpdate: func(msg Msg) Cmd {
		if msg == "work" {
			received = true
			return Quit
		}
		return nil
	}}
	p := NewProgram(m, WithContext(ctx), WithInput(nil), WithOutput(io.Discard), WithoutRenderer())
	done := make(chan struct{})
	go func() {
		p.Quit()
		close(done)
	}()
	defer func() {
		cancel()
		awaitQuitLifecycle(t, done, "pre-run Quit cleanup")
	}()
	awaitQuitLifecycle(t, done, "pre-run Quit")
	go p.Send("work")
	if _, err := p.Run(); err != nil {
		t.Fatalf("Run after an unstarted Quit failed: %v", err)
	}
	if !received {
		t.Fatal("unstarted Quit was queued and prevented later work")
	}
	p.Quit()
}

func TestProgramQuitDuringInitialization(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	initialized := make(chan struct{})
	resumeInit := make(chan struct{})
	m := quitLifecycleModel{init: func() {
		close(initialized)
		select {
		case <-resumeInit:
		case <-ctx.Done():
		}
	}}
	p := NewProgram(m, WithContext(ctx), WithInput(nil), WithOutput(io.Discard), WithoutRenderer())
	runDone := make(chan struct{})
	var runErr error
	go func() {
		_, runErr = p.Run()
		close(runDone)
	}()
	defer func() {
		cancel()
		awaitQuitLifecycle(t, runDone, "initialization Run cleanup")
	}()
	awaitQuitLifecycle(t, initialized, "model initialization")
	quitDone := make(chan struct{})
	go func() {
		p.Quit()
		close(quitDone)
	}()
	defer func() {
		cancel()
		awaitQuitLifecycle(t, quitDone, "initialization Quit cleanup")
	}()
	close(resumeInit)
	awaitQuitLifecycle(t, quitDone, "Quit during initialization")
	awaitQuitLifecycle(t, runDone, "graceful Run return")
	if runErr != nil {
		t.Fatalf("Quit during initialization failed: %v", runErr)
	}
}

func TestProgramQuitRespectsFilter(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	var rejectQuit atomic.Bool
	rejectQuit.Store(true)
	filtered := make(chan struct{}, 1)
	processed := make(chan struct{})
	m := quitLifecycleModel{onUpdate: func(msg Msg) Cmd {
		if msg == "work" {
			close(processed)
		}
		return nil
	}}
	p := NewProgram(m, WithContext(ctx), WithInput(nil), WithOutput(io.Discard), WithoutRenderer(),
		WithFilter(func(_ Model, msg Msg) Msg {
			if _, ok := msg.(QuitMsg); ok && rejectQuit.Load() {
				filtered <- struct{}{}
				return nil
			}
			return msg
		}))
	runDone := make(chan struct{})
	var runErr error
	go func() {
		_, runErr = p.Run()
		close(runDone)
	}()
	defer func() {
		cancel()
		awaitQuitLifecycle(t, runDone, "filtered Run cleanup")
	}()
	go p.Send("work")
	awaitQuitLifecycle(t, processed, "active program")
	p.Quit()
	awaitQuitLifecycle(t, filtered, "rejected Quit")
	if err := p.ctx.Err(); err != nil {
		t.Fatalf("filtered Quit canceled the program: %v", err)
	}
	rejectQuit.Store(false)
	p.Quit()
	awaitQuitLifecycle(t, runDone, "accepted Quit")
	if runErr != nil {
		t.Fatalf("accepted Quit failed: %v", runErr)
	}
}

func TestProgramSendBeforeRunStillWaits(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	var received bool
	m := quitLifecycleModel{onUpdate: func(msg Msg) Cmd {
		if msg == "work" {
			received = true
			return Quit
		}
		return nil
	}}
	p := NewProgram(m, WithContext(ctx), WithInput(nil), WithOutput(io.Discard), WithoutRenderer())
	sending := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		close(sending)
		p.Send("work")
		close(sent)
	}()
	defer func() {
		cancel()
		awaitQuitLifecycle(t, sent, "Send cleanup")
	}()
	awaitQuitLifecycle(t, sending, "pre-run Send")
	select {
	case <-sent:
		t.Fatal("Send returned before the program started")
	case <-time.After(50 * time.Millisecond):
	}
	if _, err := p.Run(); err != nil {
		t.Fatalf("Run with pre-start Send failed: %v", err)
	}
	awaitQuitLifecycle(t, sent, "delivered Send")
	if !received {
		t.Fatal("Send discarded the pre-start message")
	}
}
