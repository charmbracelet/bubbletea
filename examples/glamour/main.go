package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/muesli/termenv"
)

const content = `
Today’s Menu

| Name        | Price | Notes                           |
| ---         | ---   | ---                             |
| Tsukemono   | $2    | Just an appetizer               |
| Tomato Soup | $4    | Made with San Marzano tomatoes  |
| Okonomiyaki | $4    | Takes a few minutes to make     |
| Curry       | $3    | We can add squash if you’d like |

Seasonal Dishes

| Name                 | Price | Notes              |
| ---                  | ---   | ---                |
| Steamed bitter melon | $2    | Not so bitter      |
| Takoyaki             | $3    | Fun to eat         |
| Winter squash        | $3    | Today it's pumpkin |

Desserts

| Name         | Price | Notes                 |
| ---          | ---   | ---                   |
| Dorayaki     | $4    | Looks good on rabbits |
| Banana Split | $5    | A classic             |
| Cream Puff   | $3    | Pretty creamy!        |

All our dishes are made in-house by Karen, our chef. Most of our ingredients
are from are our garden, or the fish market down the street.

Some famous people that have eaten here lately:

* [x] René Redzepi
* [x] David Chang
* [ ] Jiro Ono (maybe some day)

Bon appétit!
`

var term = termenv.ColorProfile()

type example struct {
	viewport viewport.Model
}

func newExample() (*example, error) {
	vp := viewport.Model{Width: 78, Height: 20}

	renderer, err := glamour.NewTermRenderer(glamour.WithStylePath("notty"))
	if err != nil {
		return nil, err
	}

	str, err := renderer.Render(content)
	if err != nil {
		return nil, err
	}

	vp.SetContent(str)

	return &example{
		viewport: vp,
	}, nil
}

func (e example) Init() tea.Cmd {
	return nil
}

func (e example) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.viewport.Width = msg.Width
		return e, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return e, tea.Quit
		default:
			vp, cmd := e.viewport.Update(msg)
			e.viewport = vp
			return e, cmd
		}
	default:
		return e, nil
	}
}

func (e example) View() string {
	return e.viewport.View() + e.helpView()
}

func (e example) helpView() string {
	return termenv.String("\n  ↑/↓: Navigate • q: Quit\n").Foreground(term.Color("241")).String()
}

func main() {
	model, err := newExample()
	if err != nil {
		fmt.Println("Could not intialize Bubble Tea model:", err)
		os.Exit(1)
	}

	if err := tea.NewProgram(model).Start(); err != nil {
		fmt.Println("Bummer, there's been an error:", err)
		os.Exit(1)
	}
}
