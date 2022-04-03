# Custom Layouts

You can create your own layouts built-up of sub-components in Bubble Tea. 
To stack them horizontally or vertically, you can use [lipgloss](https://github.com/charmbracelet/lipgloss/)

You can use `lipgloss.JoinVertical` or `lipgloss.JoinHorizontal` to join sub-components in your TUI.
```go
func (m model) View() string {
	// Stack vertically
	//	s := m.a.View() + "\n\n"
	//	s += m.b.View()
	// Which is the same as 
	s := lipgloss.JoinVertical(lipgloss.Left, m.a.View(), m.b.View())
	return s
}
```
You can learn more about what you can do with Lipgloss in our [Go Docs](https://pkg.go.dev/github.com/charmbracelet/lipgloss)

## Bubbletea Sub-Components

Implementing sub-components is super common in Bubble Tea. In fact, all of the packages in Bubbles, the component library, are merely sub-components.

Nesting components in your model will be easier to manage than replacing the entire model in `Update`. It would look something like this:

```go
type model struct {
	// Sub-models
	a tea.Model
	b tea.Model
}

func (m model) Init() tea.Cmd {
	// Initialize sub-models
	return tea.Batch(
		m.a.Init(),
		m.b.Init(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// Handle any top-level messages
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			// Send the quit command. But also notice that we're not returning
			// here because in this example case we also want the sub-models to
			// do their updates too.
			cmds = append(cmds, tea.Quit)
		}
	}

	// Handle updates for sub-model A
	m.a, cmd = m.a.Update(msg)
	cmds = append(cmds, cmd)

	// Handle updates for sub-model B
	m.b, cmd = m.b.Update(msg)
	cmds = append(cmds, cmd)

	// Return updated model and any new commands
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	// Render sub model views
	s := m.a.View()
	s += m.b.View()
	return s
}
```

Of course, in a real-world scenario there would be more logic in your `Update` and `View`, but this is the gist of it.
