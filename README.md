# Tea

茶! The fun, functional way to build terminal apps. A Go framework based on
[The Elm Architecture][elm].

## Simple example

```go
package main

// A simple program that counts down from 5 and then exits.

import (
	"fmt"
	"log"
	"time"
	"github.com/charmbracelet/tea"
)

type model int

type tickMsg struct{}

func main() {
	p := tea.NewProgram(init, update, view, subscriptions)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

// Listen for messages and update the model accordingly
func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	switch msg.(type) {
	case tickMsg:
        m--
		if m <= 0 {
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

// Subscribe to events
func subscriptions(_ tea.Model) tea.Subs {
    return tea.Subs{
        "tick": tick,
    }
}

func tick(_ tea.Model) tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}
```

Hungry for more? See the [other examples][examples].

[examples]: https://github.com/charmbracelet/tea/tree/master/examples


## Authors

* [Christian Rocha](https://github.com/meowgorithm)
* [Christian Muehlhaeuser](https://github.com/muesli)

Heavily inspired by both [The Elm Architecture][elm] by Evan Czaplicki et al.
and [go-tea][gotea] by TJ Holowaychuk.

[elm]: https://guide.elm-lang.org/architecture/
[gotea]: https://github.com/tj/go-tea

## Other Resources

* [Termenv](https://github.com/muesli/termenv): advanced ANSI style and color
  support for your terminal applications. Very useful when rendering in your
  views!

## License

[MIT](https://github.com/charmbracelet/tea/raw/master/LICENSE)

***

Part of [Charm](https://charm.sh). For more info see `ssh charm.sh`.

<img alt="the Charm logo" src="https://stuff.charm.sh/charm-logotype.png" width="400px">

Charm热爱开源!
