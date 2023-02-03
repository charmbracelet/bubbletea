Commands in Bubble Tea
======================

This is the second tutorial for Bubble Tea covering commands, which deal with
I/O. The tutorial assumes you have a working knowledge of Go and a decent
understanding of [the first tutorial][basics].

You can find the non-annotated version of this program [on GitHub][source].

[basics]: https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics
[source]: https://github.com/charmbracelet/bubbletea/blob/master/tutorials/commands/main.go

## Let's Go!

For this tutorial we're building a very simple program that makes an HTTP
request to a server and reports the status code of the response.

We'll import a few necessary packages and put the URL we're going to check in
a `const`.

```go
package main

import (
    "fmt"
    "net/http"
    "os"
    "time"

    tea "github.com/charmbracelet/bubbletea"
)

const url = "https://charm.sh/"
```

## The Model

Next we'll define our model. The only things we need to store are the status
code of the HTTP response and a possible error.

```go
type model struct {
    status int
    err    error
}
```

## Commands and Messages

`Cmd`s are functions that perform some I/O and then return a `Msg`. Checking the
time, ticking a timer, reading from the disk, and network stuff are all I/O and
should be run through commands. That might sound harsh, but it will keep your
Bubble Tea program straightforward and simple.

Anyway, let's write a `Cmd` that makes a request to a server and returns the
result as a `Msg`.

```go
func checkServer() tea.Msg {

    // Create an HTTP client and make a GET request.
    c := &http.Client{Timeout: 10 * time.Second}
    res, err := c.Get(url)

    if err != nil {
        // There was an error making our request. Wrap the error we received
        // in a message and return it.
        return errMsg{err}
    }
    // We received a response from the server. Return the HTTP status code
    // as a message.
    return statusMsg(res.StatusCode)
}

type statusMsg int

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }
```

And notice that we've defined two new `Msg` types. They can be any type, even
an empty struct. We'll come back to them later in our update function.
First, let's write our initialization function.

## The Initialization Method

The initialization method is very simple: we return the `Cmd` we made earlier.
Note that we don't call the function; the Bubble Tea runtime will do that when
the time is right.

```go
func (m model) Init() (tea.Cmd) {
    return checkServer
}
```

## The Update Method

Internally, `Cmd`s run asynchronously in a goroutine. The `Msg` they return is
collected and sent to our update function for handling. Remember those message
types we made earlier when we were making the `checkServer` command? We handle
them here. This makes dealing with many asynchronous operations very easy.

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case statusMsg:
        // The server returned a status message. Save it to our model. Also
        // tell the Bubble Tea runtime we want to exit because we have nothing
        // else to do. We'll still be able to render a final view with our
        // status message.
        m.status = int(msg)
        return m, tea.Quit

    case errMsg:
        // There was an error. Note it in the model. And tell the runtime
        // we're done and want to quit.
        m.err = msg
        return m, tea.Quit

    case tea.KeyMsg:
        // Ctrl+c exits. Even with short running programs it's good to have
        // a quit key, just in case your logic is off. Users will be very
        // annoyed if they can't exit.
        if msg.Type == tea.KeyCtrlC {
            return m, tea.Quit
        }
    }

    // If we happen to get any other messages, don't do anything.
    return m, nil
}
```

## The View Function

Our view is very straightforward. We look at the current model and build a
string accordingly:

```go
func (m model) View() string {
    // If there's an error, print it out and don't do anything else.
    if m.err != nil {
        return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
    }

    // Tell the user we're doing something.
    s := fmt.Sprintf("Checking %s ... ", url)

    // When the server responds with a status, add it to the current line.
    if m.status > 0 {
        s += fmt.Sprintf("%d %s!", m.status, http.StatusText(m.status))
    }

    // Send off whatever we came up with above for rendering.
    return "\n" + s + "\n\n"
}
```

## Run the program

The only thing left to do is run the program, so let's do that! Our initial
model doesn't need any data at all in this case, we just initialize it with
as a `struct` with defaults.

```go
func main() {
    if _, err := tea.NewProgram(model{}).Run(); err != nil {
        fmt.Printf("Uh oh, there was an error: %v\n", err)
        os.Exit(1)
    }
}
```

And that's that. There's one more thing that is helpful to know about
`Cmd`s, though.

## One More Thing About Commands

`Cmd`s are defined in Bubble Tea as `type Cmd func() Msg`. So they're just
functions that don't take any arguments and return a `Msg`, which can be
any type. If you need to pass arguments to a command, you just make a function
that returns a command. For example:

```go
func cmdWithArg(id int) tea.Cmd {
    return func() tea.Msg {
        return someMsg{id: id}
    }
}
```

A more real-world example looks like:

```go
func checkSomeUrl(url string) tea.Cmd {
    return func() tea.Msg {
        c := &http.Client{Timeout: 10 * time.Second}
        res, err := c.Get(url)
        if err != nil {
            return errMsg{err}
        }
        return statusMsg(res.StatusCode)
    }
}
```

Anyway, just make sure you do as much stuff as you can in the innermost
function, because that's the one that runs asynchronously.

## Now What?

After doing this tutorial and [the previous one][basics] you should be ready to
build a Bubble Tea program of your own. We also recommend that you look at the
Bubble Tea [example programs][examples] as well as [Bubbles][bubbles],
a component library for Bubble Tea.

And, of course, check out the [Go Docs][docs].

[bubbles]: https://github.com/charmbracelet/bubbles
[docs]: https://pkg.go.dev/github.com/charmbracelet/bubbletea?tab=doc
[examples]: https://github.com/charmbracelet/bubbletea/tree/master/examples

## Additional Resources

* [Libraries we use with Bubble Tea](https://github.com/charmbracelet/bubbletea/#libraries-we-use-with-bubble-tea)
* [Bubble Tea in the Wild](https://github.com/charmbracelet/bubbletea/#bubble-tea-in-the-wild)

### Feedback

We'd love to hear your thoughts on this tutorial. Feel free to drop us a note!

* [Twitter](https://twitter.com/charmcli)
* [The Fediverse](https://mastodon.social/@charmcli)
* [Discord](https://charm.sh/chat)

***

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400"></a>

Charm热爱开源 • Charm loves open source
