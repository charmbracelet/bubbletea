// Package paginator provides a Bubble Tea package for calculating pagination
// and rendering pagination info. Note that this package does not render actual
// pages: it's purely for handling keystrokes related to pagination, and
// rendering pagination status.
package paginator

import (
	"fmt"

	tea "github.com/rprtr258/bubbletea"
	"github.com/rprtr258/bubbletea/bubbles/key"
)

// Type specifies the way we render pagination.
type Type int

// Pagination rendering options.
const (
	Arabic Type = iota
	Dots
)

// KeyMap is the key bindings for different actions within the paginator.
type KeyMap struct {
	PrevPage key.Binding
	NextPage key.Binding
}

// DefaultKeyMap is the default set of key bindings for navigating and acting
// upon the paginator.
var DefaultKeyMap = KeyMap{
	PrevPage: key.NewBinding(key.WithKeys("pgup", "left", "h")),
	NextPage: key.NewBinding(key.WithKeys("pgdown", "right", "l")),
}

// Model is the Bubble Tea model for this user interface.
type Model struct {
	// Type configures how the pagination is rendered (Arabic, Dots).
	Type Type
	// Page is the current page number.
	Page int
	// PerPage is the number of items per page.
	PerPage int
	// TotalPages is the total number of pages.
	TotalPages int
	// ActiveDot is used to mark the current page under the Dots display type.
	ActiveDot string
	// InactiveDot is used to mark inactive pages under the Dots display type.
	InactiveDot string
	// ArabicFormat is the printf-style format to use for the Arabic display type.
	ArabicFormat string

	// KeyMap encodes the keybindings recognized by the widget.
	KeyMap KeyMap

	// Deprecated: customize [KeyMap] instead.
	UsePgUpPgDownKeys bool
	// Deprecated: customize [KeyMap] instead.
	UseLeftRightKeys bool
	// Deprecated: customize [KeyMap] instead.
	UseUpDownKeys bool
	// Deprecated: customize [KeyMap] instead.
	UseHLKeys bool
	// Deprecated: customize [KeyMap] instead.
	UseJKKeys bool
}

// SetTotalPages is a helper function for calculating the total number of pages
// from a given number of items. Its use is optional since this pager can be
// used for other things beyond navigating sets. Note that it both returns the
// number of total pages and alters the model.
func (m *Model) SetTotalPages(items int) int {
	if items < 1 {
		return m.TotalPages
	}
	n := items / m.PerPage
	if items%m.PerPage > 0 {
		n++
	}
	m.TotalPages = n
	return n
}

// ItemsOnPage is a helper function for returning the number of items on the
// current page given the total number of items passed as an argument.
func (m Model) ItemsOnPage(totalItems int) int {
	if totalItems < 1 {
		return 0
	}
	start, end := m.GetSliceBounds(totalItems)
	return end - start
}

// GetSliceBounds is a helper function for paginating slices. Pass the length
// of the slice you're rendering and you'll receive the start and end bounds
// corresponding to the pagination. For example:
//
//	bunchOfStuff := []stuff{...}
//	start, end := model.GetSliceBounds(len(bunchOfStuff))
//	sliceToRender := bunchOfStuff[start:end]
func (m *Model) GetSliceBounds(length int) (start int, end int) {
	start = m.Page * m.PerPage
	end = min(m.Page*m.PerPage+m.PerPage, length)
	return start, end
}

// PrevPage is a helper function for navigating one page backward. It will not
// page beyond the first page (i.e. page 0).
func (m *Model) PrevPage() {
	if m.Page > 0 {
		m.Page--
	}
}

// NextPage is a helper function for navigating one page forward. It will not
// page beyond the last page (i.e. totalPages - 1).
func (m *Model) NextPage() {
	if !m.OnLastPage() {
		m.Page++
	}
}

// OnLastPage returns whether or not we're on the last page.
func (m Model) OnLastPage() bool {
	return m.Page == m.TotalPages-1
}

// New creates a new model with defaults.
func New() Model {
	return Model{
		Type:         Arabic,
		Page:         0,
		PerPage:      1,
		TotalPages:   1,
		KeyMap:       DefaultKeyMap,
		ActiveDot:    "•",
		InactiveDot:  "○",
		ArabicFormat: "%d/%d",
	}
}

// NewModel creates a new model with defaults.
//
// Deprecated: use [New] instead.
var NewModel = New

// Update is the Tea update function which binds keystrokes to pagination.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MsgKey:
		switch {
		case key.Matches(msg, m.KeyMap.NextPage):
			m.NextPage()
		case key.Matches(msg, m.KeyMap.PrevPage):
			m.PrevPage()
		}
	}

	return m, nil
}

// View renders the pagination to a string.
func (m Model) View() string {
	switch m.Type {
	case Dots:
		return m.dotsView()
	default:
		return m.arabicView()
	}
}

func (m Model) dotsView() string {
	var s string
	for i := 0; i < m.TotalPages; i++ {
		if i == m.Page {
			s += m.ActiveDot
			continue
		}
		s += m.InactiveDot
	}
	return s
}

func (m Model) arabicView() string {
	return fmt.Sprintf(m.ArabicFormat, m.Page+1, m.TotalPages)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
