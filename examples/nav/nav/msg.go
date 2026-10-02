package nav

import (
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
)

type PopNavigationMsg struct{}
type PushNavigationMsg tea.Model
type NoMorePageMsg struct{}

func Push(m tea.Model) tea.Cmd {
	return postCmd(PushNavigationMsg(m))
}

func Back() tea.Cmd {
	return postCmd(PopNavigationMsg{})
}

func Histories() []tea.Model {
	return nav.histories
}

func CurrentPage() tea.Model {
	return nav.curr
}

func IsActivePage(m tea.Model) bool {
	curr := nav.curr
	if curr == nil || m == nil {
		return false
	}

	return reflect.TypeOf(m).String() == reflect.TypeOf(curr).String()
}

func PushOrBack(m tea.Model) tea.Cmd {

	// 如果当前无东西, 直接返回
	if IsActivePage(m) {
		return Back()
	}

	return Push(m)
}

func postCmd(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
