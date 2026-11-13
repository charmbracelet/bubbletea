package nav

import (
	tea "github.com/charmbracelet/bubbletea"
)

// PageLife  will call every time.
type PageLife interface {
	OnEntering() (tea.Model, tea.Cmd)
	OnLeaving() (tea.Model, tea.Cmd)
}
