package list

import (
	"fmt"
	"io"
	"testing"

	tea "github.com/rprtr258/bubbletea"
	"github.com/stretchr/testify/assert"
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                          { return 1 }
func (d itemDelegate) Spacing() int                         { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m Model, index int, listItem Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	fmt.Fprint(w, m.Styles.TitleBar.Render(fmt.Sprintf("%d. %s", index+1, i)))
}

func TestStatusBarItemName(t *testing.T) {
	list := New([]Item{item("foo"), item("bar")}, itemDelegate{}, 10, 10)
	assert.Contains(t, list.statusView(), "2 items")

	list.SetItems([]Item{item("foo")})
	assert.Contains(t, list.statusView(), "1 item")
}

func TestStatusBarWithoutItems(t *testing.T) {
	list := New([]Item{}, itemDelegate{}, 10, 10)
	assert.Contains(t, list.statusView(), "No items")
}

func TestCustomStatusBarItemName(t *testing.T) {
	list := New([]Item{item("foo"), item("bar")}, itemDelegate{}, 10, 10)
	list.SetStatusBarItemName("connection", "connections")
	assert.Contains(t, list.statusView(), "2 connections")

	list.SetItems([]Item{item("foo")})
	assert.Contains(t, list.statusView(), "1 connection")

	list.SetItems([]Item{})
	assert.Contains(t, list.statusView(), "No connections")
}
