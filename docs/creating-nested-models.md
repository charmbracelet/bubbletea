# Creating Nested Models

There may be situations where you want to have your own nested models for your project. 
This can allow you to toggle between different views and organize your logic for `Update`.

## Showing Specific Nested Models
In Bubble Tea programs, you can decide which components are shown by holding a `state` field in your main model struct. 
For example:

```go
// this is an enum for Go
var sessionState uint
const (
	timerView sessionState = iota
	statsView
)

type mainModel struct {
	state sessionState
}

func New() mainModel {
	// initialize your model; timerView is the first "view" we want to see
	return mainModel{state: timerView}
}

func (m mainModel) Init() tea.Cmd {
	return nil
}

func (m mainModel) Update(msg tea.Msg) {
	switch msg := msg.(type) {
		// Handle IO -> keypress, WindowSizeMSg
		case tea.KeyMsg:
			return m, tea.Quit
		case tea.WindowSizeMSg:
			// handle resizing windows
		// handle your Msgs
	}
}

func (m mainModel) View() string {
	switch m.state {
		case statsView:
			return statsView.View()
		default:
			return "timer is " + timerView.View()
	}
}
```
As you can see, the main impact the `state` has is in the `View` method.
