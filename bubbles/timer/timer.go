// Package timer provides a simple timeout component.
package timer

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

// Authors note with regard to start and stop commands:
//
// Technically speaking, sending commands to start and stop the timer in this
// case is extraneous. To stop the timer we'd just need to set the 'running'
// property on the model to false which cause logic in the update function to
// stop responding to TickMsgs. To start the model we'd set 'running' to true
// and fire off a TickMsg. Helper functions would look like:
//
//     func (m *model) Start() tea.Cmd
//     func (m *model) Stop()
//
// The danger with this approach, however, is that order of operations becomes
// important with helper functions like the above. Consider the following:
//
//     // Would not work
//     return m, m.timer.Start()
//
//	   // Would work
//     cmd := m.timer.start()
//     return m, cmd
//
// Thus, because of potential pitfalls like the ones above, we've introduced
// the extraneous StartStopMsg to simplify the mental model when using this
// package. Bear in mind that the practice of sending commands to simply
// communicate with other parts of your application, such as in this package,
// is still not recommended.

// StartStopMsg is used to start and stop the timer.
type StartStopMsg struct {
	ID      int
	running bool
}

// TickMsg is a message that is sent on every timer tick.
type TickMsg struct {
	// ID is the identifier of the timer that sends the message. This makes
	// it possible to determine which timer a tick belongs to when there
	// are multiple timers running.
	//
	// Note, however, that a timer will reject ticks from other timers, so
	// it's safe to flow all TickMsgs through all timers and have them still
	// behave appropriately.
	ID int

	// Timeout returns whether or not this tick is a timeout tick. You can
	// alternatively listen for TimeoutMsg.
	Timeout bool
}

// TimeoutMsg is a message that is sent once when the timer times out.
//
// It's a convenience message sent alongside a TickMsg with the Timeout value
// set to true.
type TimeoutMsg struct {
	ID int
}

// Model of the timer component.
type Model struct {
	// How long until the timer expires.
	Timeout time.Duration

	// How long to wait before every tick. Defaults to 1 second.
	Interval time.Duration

	id      int
	running bool
}

// NewWithInterval creates a new timer with the given timeout and tick interval.
func NewWithInterval(timeout, interval time.Duration) Model {
	return Model{
		Timeout:  timeout,
		Interval: interval,
		running:  true,
		id:       nextID(),
	}
}

// New creates a new timer with the given timeout and default 1s interval.
func New(timeout time.Duration) Model {
	return NewWithInterval(timeout, time.Second)
}

// ID returns the model's identifier. This can be used to determine if messages
// belong to this timer instance when there are multiple timers.
func (m Model) ID() int {
	return m.id
}

// Running returns whether or not the timer is running. If the timer has timed
// out this will always return false.
func (m Model) Running() bool {
	if m.Timedout() || !m.running {
		return false
	}
	return true
}

// Timedout returns whether or not the timer has timed out.
func (m Model) Timedout() bool {
	return m.Timeout <= 0
}

// Init starts the timer.
func (m Model) Init() tea.Cmd {
	return m.tick()
}

// Update handles the timer tick.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StartStopMsg:
		if msg.ID != 0 && msg.ID != m.id {
			return m, nil
		}
		m.running = msg.running
		return m, m.tick()
	case TickMsg:
		if !m.Running() || (msg.ID != 0 && msg.ID != m.id) {
			break
		}

		m.Timeout -= m.Interval
		return m, tea.Batch(m.tick(), m.timedout())
	}

	return m, nil
}

// View of the timer component.
func (m Model) View() string {
	return m.Timeout.String()
}

// Start resumes the timer. Has no effect if the timer has timed out.
func (m *Model) Start() tea.Cmd {
	return m.startStop(true)
}

// Stop pauses the timer. Has no effect if the timer has timed out.
func (m *Model) Stop() tea.Cmd {
	return m.startStop(false)
}

// Toggle stops the timer if it's running and starts it if it's stopped.
func (m *Model) Toggle() tea.Cmd {
	return m.startStop(!m.Running())
}

func (m Model) tick() tea.Cmd {
	return tea.Tick(m.Interval, func(_ time.Time) tea.Msg {
		return TickMsg{ID: m.id, Timeout: m.Timedout()}
	})
}

func (m Model) timedout() tea.Cmd {
	if !m.Timedout() {
		return nil
	}
	return func() tea.Msg {
		return TimeoutMsg{ID: m.id}
	}
}

func (m Model) startStop(v bool) tea.Cmd {
	return func() tea.Msg {
		return StartStopMsg{ID: m.id, running: v}
	}
}
