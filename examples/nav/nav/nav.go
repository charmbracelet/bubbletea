package nav

import (
	tea "github.com/charmbracelet/bubbletea"
)

var (
	nav = &navigator{}
)

type navigator struct {
	histories []tea.Model
	curr      tea.Model
}

func View() string {
	if nav.curr == nil {
		return ""
	}

	return nav.curr.View()
}

func Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case PushNavigationMsg:
		return handlePushNavigation(msg)
	case PopNavigationMsg:
		return handlePopNavigation()
	}

	if nav.curr == nil {
		return nil
	}

	// update page
	var c tea.Cmd
	nav.curr, c = nav.curr.Update(msg)
	return c
}

// handlePushNavigation 处理页面压入导航栈的操作
func handlePushNavigation(msg PushNavigationMsg) tea.Cmd {
	var (
		cmds []tea.Cmd
		c    tea.Cmd
	)

	// 1. 处理当前页面的离开逻辑
	if nav.curr != nil {
		// 将当前页面存入历史记录
		nav.histories = append(nav.histories, nav.curr)

		// 调用当前页面的离开钩子
		if p, ok := nav.curr.(PageLife); ok {
			nav.curr, c = p.OnLeaving()
			cmds = append(cmds, c)
		}
	}

	// 2. 压入新页面
	nav.curr = tea.Model(msg)
	// 3. 初始化新页面并处理进入逻辑
	cmds = append(cmds, nav.curr.Init())

	// 页面离开钩子
	if p, ok := nav.curr.(PageLife); ok {
		nav.curr, c = p.OnEntering()
		cmds = append(cmds, c)
	}

	return tea.Sequence(cmds...)
}

// handlePopNavigation 处理页面弹出导航栈的操作
func handlePopNavigation() tea.Cmd {

	// 尝试弹出页面
	if nav.curr == nil {
		return postCmd(NoMorePageMsg{})
	}

	if len(nav.histories) == 0 {
		return postCmd(NoMorePageMsg{})
	}

	var (
		cmds []tea.Cmd
		c    tea.Cmd
	)
	// 1. 处理旧页面的离开逻辑
	if p, ok := nav.curr.(PageLife); ok {
		_, c = p.OnLeaving()
		cmds = append(cmds, c)
	}

	// 2. 处理新页面的进入逻辑
	nav.curr = nav.histories[len(nav.histories)-1]
	nav.histories = nav.histories[:len(nav.histories)-1]

	if p, ok := nav.curr.(PageLife); ok {
		m, cmd := p.OnEntering()
		nav.curr = m
		cmds = append(cmds, cmd)
	}

	return tea.Sequence(cmds...)
}
