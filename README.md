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
	err := tea.NewProgram(model(5), update, view, []tea.Sub{tick}).Start()
	if err != nil {
		log.Fatal(err)
	}
}

func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	switch msg.(type) {
	case tickMsg:
		m -= 1
		if m <= 0 {
			return m, tea.Quit
		}
	}
	return m, nil
}

func view(mdl tea.Model) string {
	m, _ := mdl.(model)
	return fmt.Sprintf("Hi. This program will exit in %d seconds...\n", m)
}

func tick(_ tea.Model) tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}
```

Hungry for more? See the [other examples][examples].

[examples]: https://github.com/charmbracelet/tea/examples

## Credit

Heavily inspired by both [The Elm Architecture][elm] by Evan Czaplicki et al.
and [go-tea][gotea] by TJ Holowaychuk.

[elm]: https://guide.elm-lang.org/architecture/
[gotea]: https://github.com/tj/go-tea

***

Part of [Charm](https://charm.sh). For more info see `ssh charm.sh`. Charm热爱开源!
