// package paginator provides a Bubble Tea package for calulating pagination
// and rendering pagination info. Note that this package does not render actual
// pages: it's purely for handling keystrokes related to pagination, and
// rendering pagination status.
package paginator

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Type specifies the way we render pagination.
type Type int

// Pagination rendering options
const (
	Arabic Type = iota
	Dots
)

// Model is the Tea model for this user interface.
type Model struct {
	Type             Type
	Page             int
	PerPage          int
	TotalPages       int
	ActiveDot        string
	InactiveDot      string
	ArabicFormat     string
	UseLeftRightKeys bool
	UseUpDownKeys    bool
	UseHLKeys        bool
	UseJKKeys        bool
}

// SetTotalPages is a helper function for calculatng the total number of pages
// from a given number of items. It's use is optional since this pager can be
// used for other things beyond navigating sets. Note that it both returns the
// number of total pages and alters the model.
func (m *Model) SetTotalPages(items int) int {
	if items == 0 {
		return 0
	}
	n := items / m.PerPage
	if items%m.PerPage > 0 {
		n++
	}
	m.TotalPages = n
	return n
}

// ItemsOnPage is a helper function for returning the numer of items on the
// current page given the total numer of items passed as an argument.
func (m Model) ItemsOnPage(totalItems int) int {
	start, end := m.GetSliceBounds(totalItems)
	return end - start
}

// GetSliceBounds is a helper function for paginating slices. Pass the length
// of the slice you're rendering and you'll receive the start and end bounds
// corresponding the to pagination. For example:
//
//     bunchOfStuff := []stuff{...}
//     start, end := model.GetSliceBounds(len(bunchOfStuff))
//     sliceToRender := bunchOfStuff[start:end]
//
func (m *Model) GetSliceBounds(length int) (start int, end int) {
	start = m.Page * m.PerPage
	end = min(m.Page*m.PerPage+m.PerPage, length)
	return start, end
}

// PrevPage is a number function for navigating one page backward. It will not
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

// LastPage returns whether or not we're on the last page.
func (m Model) OnLastPage() bool {
	return m.Page == m.TotalPages-1
}

// NewModel creates a new model with defaults.
func NewModel() Model {
	return Model{
		Type:             Arabic,
		Page:             0,
		PerPage:          1,
		TotalPages:       1,
		ActiveDot:        "•",
		InactiveDot:      "○",
		ArabicFormat:     "%d/%d",
		UseLeftRightKeys: true,
		UseUpDownKeys:    false,
		UseHLKeys:        true,
		UseJKKeys:        false,
	}
}

// Update is the Tea update function which binds keystrokes to pagination.
func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UseLeftRightKeys {
			switch msg.String() {
			case "left":
				m.PrevPage()
			case "right":
				m.NextPage()
			}
		}
		if m.UseUpDownKeys {
			switch msg.String() {
			case "up":
				m.PrevPage()
			case "down":
				m.NextPage()
			}
		}
		if m.UseHLKeys {
			switch msg.String() {
			case "h":
				m.PrevPage()
			case "l":
				m.NextPage()
			}
		}
		if m.UseJKKeys {
			switch msg.String() {
			case "j":
				m.PrevPage()
			case "k":
				m.NextPage()
			}
		}
	}

	return m, nil
}

// View renders the pagination to a string.
func View(model tea.Model) string {
	m, ok := model.(Model)
	if !ok {
		return "could not perform assertion on model"
	}
	switch m.Type {
	case Dots:
		return dotsView(m)
	default:
		return arabicView(m)
	}
}

func dotsView(m Model) string {
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

func arabicView(m Model) string {
	return fmt.Sprintf(m.ArabicFormat, m.Page+1, m.TotalPages)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
