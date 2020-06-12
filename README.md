# Bubble Tea

The fun, functional way to build terminal apps. A Go framework based on
[The Elm Architecture][elm].

⚠️  This project is a pre-release so the API is subject to change
a little. That said, we're using it in production.


## Simple example

```go
package main

// A simple program that counts down from 5 and then exits.

import (
	"fmt"
	"log"
	"time"
	tea "github.com/charmbracelet/bubbletea"
)

type model int

type tickMsg struct{}

func main() {
	p := tea.NewProgram(initialize, update, view, subscriptions)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

// Return the initial model and initial command
func initialize() (tea.Model, tea.Cmd) {
    return 5, tick
}

// Listen for messages and update the model accordingly
func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	switch msg.(type) {
	case tickMsg:
		m--
		if m == 0 {
			return m, tea.Quit
		}
	}
	return m, nil
}

// Render to the terminal
func view(mdl tea.Model) string {
	m, _ := mdl.(model)
	return fmt.Sprintf("Hi. This program will exit in %d seconds...\n", m)
}

// A simple command which Bubble Tea runs asynchronously.
func tick() tea.Msg {
    time.Sleep(time.Second)
    return tickMsg{}
}
```

Hungry for more? Totally confused? See the [other examples][examples].

[examples]: https://github.com/charmbracelet/tea/tree/master/examples


## Other Resources

* [termenv](https://github.com/muesli/termenv): advanced ANSI style and color
  support for your terminal applications. Very useful when rendering your
  views.
* [reflow](https://github.com/muesli/reflow): a collection of ANSI-aware text
  formatting tools. Also useful for view rendering.
* [go-runewidth](https://github.com/mattn/go-runewidth): functions to get the
  physical width of runes in terms of cells. Indispensable when working with
  fullwidth and zero-width characters.


## Acknowledgments

Heavily inspired by both [The Elm Architecture][elm] by Evan Czaplicki et al.
and [go-tea][gotea] by TJ Holowaychuk.

[elm]: https://guide.elm-lang.org/architecture/
[gotea]: https://github.com/tj/go-tea


## License

[MIT](https://github.com/charmbracelet/tea/raw/master/LICENSE)

***

Part of [Charm](https://charm.sh).

<img alt="the Charm logo" src="https://stuff.charm.sh/charm-logotype.png" width="400px">

Charm热爱开源!
