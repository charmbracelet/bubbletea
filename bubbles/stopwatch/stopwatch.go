// Package stopwatch provides a simple stopwatch component.
package stopwatch

import (
	"sync"
	"time"

	tea "github.com/rprtr258/bubbletea"
)

var (
	lastID int
	idMtx  sync.Mutex
)

func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

// TickMsg is a message that is sent on every timer tick.
type TickMsg struct {
	// ID is the identifier of the stopwatch that sends the message. This makes
	// it possible to determine which stopwatch a tick belongs to when there
	// are multiple stopwatches running.
	//
	// Note, however, that a stopwatch will reject ticks from other
	// stopwatches, so it's safe to flow all TickMsgs through all stopwatches
	// and have them still behave appropriately.
	ID int
}

// StartStopMsg is sent when the stopwatch should start or stop.
type StartStopMsg struct {
	ID      int
	running bool
}

// ResetMsg is sent when the stopwatch should reset.
type ResetMsg struct {
	ID int
}

// Model for the stopwatch component.
type Model struct {
	d       time.Duration
	id      int
	running bool

	// How long to wait before every tick. Defaults to 1 second.
	Interval time.Duration
}

// NewWithInterval creates a new stopwatch with the given timeout and tick
// interval.
func NewWithInterval(interval time.Duration) Model {
	return Model{
		Interval: interval,
		id:       nextID(),
	}
}

// New creates a new stopwatch with 1s interval.
func New() Model {
	return NewWithInterval(time.Second)
}

// ID returns the unique ID of the model.
func (m Model) ID() int {
	return m.id
}

// Init starts the stopwatch.
func (m Model) Init() tea.Cmd {
	return m.Start()
}

// Start starts the stopwatch.
func (m Model) Start() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		return StartStopMsg{ID: m.id, running: true}
	}, tick(m.id, m.Interval))
}

// Stop stops the stopwatch.
func (m Model) Stop() tea.Cmd {
	return func() tea.Msg {
		return StartStopMsg{ID: m.id, running: false}
	}
}

// Toggle stops the stopwatch if it is running and starts it if it is stopped.
func (m Model) Toggle() tea.Cmd {
	if m.Running() {
		return m.Stop()
	}
	return m.Start()
}

// Reset resets the stopwatch to 0.
func (m Model) Reset() tea.Cmd {
	return func() tea.Msg {
		return ResetMsg{ID: m.id}
	}
}

// Running returns true if the stopwatch is running or false if it is stopped.
func (m Model) Running() bool {
	return m.running
}

// Update handles the timer tick.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StartStopMsg:
		if msg.ID != m.id {
			return m, nil
		}
		m.running = msg.running
	case ResetMsg:
		if msg.ID != m.id {
			return m, nil
		}
		m.d = 0
	case TickMsg:
		if !m.running || msg.ID != m.id {
			break
		}
		m.d += m.Interval
		return m, tick(m.id, m.Interval)
	}

	return m, nil
}

// Elapsed returns the time elapsed.
func (m Model) Elapsed() time.Duration {
	return m.d
}

// View of the timer component.
func (m Model) View() string {
	return m.d.String()
}

func tick(id int, d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return TickMsg{ID: id}
	})
}
