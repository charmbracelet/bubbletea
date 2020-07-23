# Bubble Tea Tutorial

Bubble Tea is based on the functional design paradigms of [The Elm
Architecture][elm]. It might not seem very Go-like at first, but once you get
used to the general structure you'll find that most of the idomatic Go things
you know and love are still relevant and useful here.

By the way, the non-annotated version of of this program is available
[on GitHub](https://github.com/charmbracelet/bubbletea/master/tutorials/basics).

This tutorial assumes you have a working knowledge of Go.

[elm]: https://guide.elm-lang.org/architecture/

## Enough! Let's get to it.

For this tutorial we're making a to-do list.

To start we'll define our package and import some libraries. Our only external
import will be the Bubble Tea, library, which we'll call `tea` for short.

```go
    package main

    import (
        "fmt"
        "os"

        tea "github.com/charmbracelet/bubbletea"
    )
```

Bubble Tea programs are comprised of a model that describes the application
state and three simple functions that are centered around the model:

* **Initialize**, a function that returns the model's initial state.
* **Update**, a function that handles incoming events and updates the model accordingly.
* **View**, a function that renders the UI based on the data in the model.

## The Model

So let's start by defining our application's model. The model is simply the
application's state. It can be any type, but a `struct` usually makes the most
sense.

```go
    type model struct {
        choices  []string           // items on the to-do list
        cursor   int                // which to-do list item our cursor is pointing at
        selected map[int]struct{}   // which to-do items are selected
    }
```

## Initialize

Next we'll define a function that will initialize our application. An
initialize function returns a model representing our application's initial
state, as well as a `Cmd` that could perform some initial I/O. For now, we
don't need to do any I/O, so for the command we'll just return nil, which
translate to "no command."

```go
    func initialize() (tea.Model, tea.Cmd) {
        m := model{

            // Our to-do list is just a grocery list
            choices:  []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},

            // A map which indicates which choices are selected. We're using
            // the  map like a mathematical set. The keys refer to the indexes
            // of the `choices` slice, above.
            selected: make(map[int]struct{}),
        }

        // Return the model and `nil`, which means "no I/O right now, please."
        return m, nil
    }
```

## Update

Next we'll define the update function. The update function is called when
"things happen." It's job is to look at what has happened and return an
updated model based on whatever happened. It can also return a `Cmd` and make
more things happen, but we'll get into that later.

In our case, when a user presses the down arrow `update`'s job is to notice
that the down arrow was pressed and move the cursor accordingly (or not).

The "something happened" comes in as a `Msg`, which can be any type. Messages
indicate some I/O happened, such as a keypress, timer, or a response from
a server.

We usually figure out which type of `Msg` we received with a type switch, but
you could also use a type assertion.

For now, we'll just deal with `tea.KeyMsg`, which are automatically sent to
the update function when keys are pressed.

```go
    func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
        m, _ := mdl.(model)

        switch msg := msg.(type) {

        // Is it a key press?
        case tea.KeyMsg:

            // Cool, what was the actual key pressed?
            switch msg.String() {

            // These keys should exit the program.
            case "ctrl+c", "q":
                return m, tea.Quit

            // The "up" and "k" keys move the cursor up
            case "up", "k":
                if m.cursor > 0 {
                    m.cursor--
                }

            // The "down" and "j" keys move the cursor down
            case "down", "j":
                if m.cursor < len(m.choices)-1 {
                    m.cursor++
                }

            // The "enter" key and the spacebar (a literal space) toggle
            // the selected state for the item that the cursor is pointing at.
            case "enter", " ":
                _, ok := m.selected[m.cursor]
                if ok {
                    delete(m.selected, m.cursor)
                } else {
                    m.selected[m.cursor] = struct{}{}
                }
            }
        }

        // Return the updated model to the Bubble Tea runtime for processing.
        // Note that we're not returning a command.
        return m, nil
    }
```

You may have noticed that "ctrl+c" and "q" above return a `tea.Quit` with the
model. That's a special command which instructs the Bubble Tea runtime to exit,
effectively quitting the program.

## The View

At last, it's time to render our UI. Of all the functions, the view is the
simplest. A model, in it's current state, comes in and a `string` comes out.
That string is our UI!

Because the view describes the entire UI of your application, you don't have
to worry about redraw logic and stuff like that. Bubble Tea takes care of it
for you.

```go
    func view(mdl tea.Model) string {
        m, _ := mdl.(model)

        // The header
        s := "What should we buy at the market?\n\n"

        // Iterate over our choices
        for i, choice := range m.choices {

            // Is the cursor pointing at this choice?
            cursor := " " // no cursor
            if m.cursor == i {
                cursor = ">" // cursor!
            }

            // Is this choice selected?
            checked := " " // not selected
            if _, ok := m.selected[i]; ok {
                checked = "x" // selected!
            }

            // Render the row
            s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
        }

        // The footer
        s += "\nPress q to quit.\n"

        // Send off the UI to rendered
        return s
    }
```

## All togeher now

The last step is to simply run our program. We pass our functions to
`tea.NewProgram` and let it rip:

```go
    func main() {
        p := tea.NewProgram(initialize, update, view)
        if err := p.Start(); err != nil {
            fmt.Printf("Alas, there's been an error: %v", err)
            os.Exit(1)
        }
    }
```

## What's next?

This tutorial covers the basics of building an interactive terminal UI, but
in the real world you'll also need to perform I/O. To learn about that have a
look at the [Cmd Tutorial][cmd]. It's pretty simple.

There are also several [examples][examples] available. Many of the examples
make use of [Bubbles][bubbles], the little Bubble Tea component library which
includes handy things like a text input component, spinners and a viewport.

Of course, there are also [Go Docs][docs] for Bubble Tea.

[cmd]: http://github.com/charmbracelet/bubbletea/tree/master/tutorials/cmds/
[examples]: http://github.com/charmbracelet/bubbletea/tree/master/examples
[bubbles]: https://github.com/charmbracelet/bubbles
[docs]: https://pkg.go.dev/github.com/charmbracelet/glow?tab=doc

## Feedback

We'd love to hear your thoughts on this tutorial. Please feel free to reach out
anytime.

* [Twitter](https://twitter.com/charmcli)
* [The Fediverse](https://mastodon.technology/@charm)
