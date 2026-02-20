package main

import (
	"context"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var (
	terminalHeight, terminalWidth int

	baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#3FFF50"))

	evenMsgStyle = baseStyle.Align(lipgloss.Center).Italic(true)
	oddMsgStyle  = evenMsgStyle.Faint(true)

	convoUsernameStyle           = baseStyle.Faint(true)
	partialSelConvoUsernameStyle = convoUsernameStyle.Faint(false)
	selConvoUsernameStyle        = partialSelConvoUsernameStyle.Underline(true).Italic(true)
)

type Model struct {
	msgVP viewport.Model
	// communication from messageHandlerService to Model
	msgChan <-chan message
	// communication from Model to messageHandlerService
	usernameChan     chan<- username
	murderMsgService context.CancelFunc
	partialSelTabIdx int
	selTabIdx        int
	msgs             []message
}

func initialModel(msgC <-chan message, userC chan<- username, cFunc context.CancelFunc) Model {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("#808080")).
		Align(lipgloss.Center)

	return Model{
		msgChan:          msgC,
		usernameChan:     userC,
		murderMsgService: cFunc,
		partialSelTabIdx: -1,
		selTabIdx:        -1,
		msgVP:            vp,
	}

}

func (m Model) Init() tea.Cmd {
	return m.listenForMsgs()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		terminalWidth = msg.Width
		terminalHeight = msg.Height
		m.updateMsgVPDimensions()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.murderMsgService()
			close(m.usernameChan)
			return m, tea.Quit

		case "tab":
			m.partialSelTabIdx = (m.partialSelTabIdx + 1) % len(dummyUsers) // move tab index between usernames

		case "enter":
			if m.partialSelTabIdx >= 0 {
				m.selTabIdx = m.partialSelTabIdx
				m.usernameChan <- dummyUsers[m.selTabIdx]
			}

		case "esc":
			m.usernameChan <- "" // no username selected
			m.selTabIdx = -1
			m.partialSelTabIdx = -1
		}

	case message:
		m.msgs = append(m.msgs, msg)
		m.setMsgVPContent() // set the updated msg slice render
		m.msgVP.LineDown(2)
		return m, tea.Batch(m.listenForMsgs(), m.handleMsgVPUpdate(msg)) // continue listening
	}

	return m, m.handleMsgVPUpdate(msg)

}

func (m Model) View() string {
	usrBtns := m.renderConvoUserBtns()
	help := lipgloss.NewStyle().Faint(true).MarginTop(1).Render("[ TAB | ENTER | ESC | CTRL + C ]")
	s := lipgloss.JoinVertical(lipgloss.Center, m.msgVP.View(), usrBtns, help)
	return lipgloss.Place(terminalWidth, terminalHeight, lipgloss.Center, lipgloss.Center, s)
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func (m *Model) updateMsgVPDimensions() {
	s := strings.Join([]string{string(dummyUsers[0]), string(dummyUsers[1]), string(dummyUsers[2]), string(dummyUsers[3])}, "   ")
	m.msgVP.Width = lipgloss.Width(s) + 4
	m.msgVP.Height = terminalHeight - 10
}

func (m *Model) handleMsgVPUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.msgVP, cmd = m.msgVP.Update(msg)
	return cmd
}

func (m *Model) setMsgVPContent() {
	var sb strings.Builder
	w := m.msgVP.Width - m.msgVP.Style.GetHorizontalFrameSize()
	for i, msg := range m.msgs {
		// using styles to create visual hierarchy
		if i%2 == 0 {
			sb.WriteString(evenMsgStyle.Width(w).Render(string(msg)))
		} else {
			sb.WriteString(oddMsgStyle.Width(w).Render(string(msg)))
		}
		sb.WriteString("\n")
	}
	m.msgVP.SetContent(sb.String())
}

func (m Model) renderConvoUserBtns() string {
	var sb strings.Builder
	for i, usrname := range dummyUsers {
		style := convoUsernameStyle
		if i == m.selTabIdx {
			style = selConvoUsernameStyle
		} else if i == m.partialSelTabIdx {
			style = partialSelConvoUsernameStyle
		}
		sb.WriteString(style.Render(string(usrname)))
		if i < len(dummyUsers)-1 { // username divider
			sb.WriteString(lipgloss.NewStyle().Faint(true).Render(" ⇋ "))
		}
	}
	return sb.String()
}

func (m Model) listenForMsgs() tea.Cmd {
	return func() tea.Msg {
		return <-m.msgChan
	}
}
