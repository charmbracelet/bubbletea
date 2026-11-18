# Common Patterns in Bubble Tea

You've become comfortable with the basics of Bubble Tea, `tea.Cmd`, and
`tea.Msg`, *but* you're still unsure if your solution is following best
practices? We get asked about this a lot, so we decided to take some time
to investigate what questions are being asked most and assess our own patterns
when building command line apps with Bubble Tea. In this tutorial, we'll
highlight some common patterns that you'll come across when building Bubble Tea
applications that should help to simplify your solutions.

## "I want multiple elements in a single view"

<img width="800" src="https://github.com/charmbracelet/bubbletea/blob/master/examples/composable-views/composable-views.gif" />

If you have a composite view, then you have multiple components on one screen
that you want to be able to switch between. To handle this in Bubble Tea, you'll
want your parent component to house a `state` field that dictates which element
on the screen is focused and receiving key presses.

You can see a [basic example][basic] of this where we switch focus between a
timer and spinner.

```go
		switch m.state {
		// update whichever model is focused
		case spinnerView:
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		default:
			m.timer, cmd = m.timer.Update(msg)
			cmds = append(cmds, cmd)
		}
```

This same strategy can be used for switching between different models. We do
just this in [Glow][glow] to switch between the file listing and viewing the
document.

## "I only want the model that triggered the message to update"

To figure out whether a component should process the message or not, simply
include an ID in the message. We'll then compare the ID fields in the message
and your child model. If they are a match, then we handle that message in the
child's `Update`, otherwise it just gets ignored.

This pattern is used in the [spinner bubble][spinner]:

<img width="800" src="https://github.com/charmbracelet/bubbletea/blob/patterns/examples/spinner/spinner.gif" />

```go
// Update is the Tea update function.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		// If an ID is set, and the ID doesn't belong to this spinner, reject
		// the message.
		if msg.ID > 0 && msg.ID != m.id {
			return m, nil
		}

		// If a tag is set, and it's not the one we expect, reject the message.
		// This prevents the spinner from receiving too many messages and
		// thus spinning too fast.
		if msg.tag > 0 && msg.tag != m.tag {
			return m, nil
		}

		m.frame++
		if m.frame >= len(m.Spinner.Frames) {
			m.frame = 0
		}

		m.tag++
		// include the ID of the model that triggered the msg
		return m, m.tick(m.id, m.tag)
	default:
		return m, nil
	}
}
```

This is what that `tick` function does:

```go
func (m Model) tick(id, tag int) tea.Cmd {
	return tea.Tick(m.Spinner.FPS, func(t time.Time) tea.Msg {
		return TickMsg{
			Time: t,
			ID:   id,
			tag:  tag,
		}
	})
}
```

[Source](https://github.com/charmbracelet/bubbles/blob/master/spinner/spinner.go#L195-L203l)

## "I want my Bubble Tea program to display external processes"

You can send information from outside processes to your Bubble Tea
applications. There are a couple of examples on how to handle this behavior in
the Bubble Tea Repo: 
- [downloading a file and feeding the progress to Bubble Tea][progress-download]
- [a `p.Send` example that simulates a message from outside the program][send-msg]. 

<img width="800" src="https://github.com/charmbracelet/bubbletea/blob/master/examples/send-msg/send-msg.gif" />

The goal here is to have the external process run in a [Goroutine][goroutine].

The steps are as follows:
1. Create a new `tea.Program` with your model.
2. Start a Goroutine for the external process you want to document in your
   Bubble Tea program.
3. Use [`p.Send`][psend] to send the data to your Bubble Tea application. 
4. Run your `tea.Program` outside the Goroutine.
5. Handle that message type in your `Update` function.

In the simpler `p.Send` example, it looks like this:
```go
func main() {
	p := tea.NewProgram(newModel())

	// Simulate activity
	go func() {
		for {
			pause := time.Duration(rand.Int63n(899)+100) * time.Millisecond // nolint:gosec
			time.Sleep(pause)

			// Send the Bubble Tea program a message from outside the
			// tea.Program. This will block until it is ready to receive
			// messages.
			p.Send(resultMsg{food: randomFood(), duration: pause})
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// ...
		case resultMsg:
		m.results = append(m.results[1:], msg)
		return m, nil
	}
}
```
[Source][send-msg]

In the more complex download example, it looks like this:

```go
func (pw *progressWriter) Start() {
	// TeeReader calls pw.Write() each time a new response is received
	_, err := io.Copy(pw.file, io.TeeReader(pw.reader, pw))
	if err != nil {
		p.Send(progressErrMsg{err})
	}
}

// ...

func main() {
	// ...
	pw := &progressWriter{
		total:  int(resp.ContentLength),
		file:   file,
		reader: resp.Body,
		onProgress: func(ratio float64) {
			p.Send(progressMsg(ratio))
		},
	}

	m := model{
		pw:       pw,
		progress: progress.New(progress.WithDefaultGradient()),
	}
	// Start Bubble Tea
	p = tea.NewProgram(m)

	// Start the download
	go pw.Start()

	if _, err := p.Run(); err != nil {
		fmt.Println("error running program:", err)
		os.Exit(1)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// ...
	case progressMsg:
		var cmds []tea.Cmd

		if msg >= 1.0 {
			cmds = append(cmds, tea.Sequence(finalPause(), tea.Quit))
		}

		cmds = append(cmds, m.progress.SetPercent(float64(msg)))
		return m, tea.Batch(cmds...)
	}
}
```
[Source][progress-download]

## Additional Resources

* [Official examples of Bubble Tea usage][examples]
* [The Bubble Tea API](https://pkg.go.dev/github.com/charmbracelet/bubbletea)
* [Libraries we use with Bubble Tea](https://github.com/charmbracelet/bubbletea/#libraries-we-use-with-bubble-tea)
* [Bubble Tea in the wild](https://github.com/charmbracelet/bubbletea/#bubble-tea-in-the-wild)

And don't forget to check out the other [tutorials][tutorials] if you're just
getting started with Bubble Tea.

### Feedback

Let us know in [GitHub discussions][discuss] if there are other patterns that
you'd like to see! If there's enough interest we can certainly include it.
Don't be shy, we love to hear from you.

* [Twitter](https://twitter.com/charmcli)
* [The Fediverse](https://mastodon.social/@charmcli)
* [Discord](https://charm.sh/chat)

***

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400"></a>

Charm热爱开源 • Charm loves open source


[discuss]: https://github.com/charmbracelet/bubbletea/discussions
[tutorials]: https://github.com/charmbracelet/bubbletea/tree/patterns/tutorials
[examples]: https://github.com/charmbracelet/bubbletea/tree/master/examples
[psend]: https://pkg.go.dev/github.com/charmbracelet/bubbletea#Program.Send
[goroutine]: https://go.dev/doc/effective_go#goroutines
[send-msg]: https://github.com/charmbracelet/bubbletea/blob/master/examples/send-msg/main.go
[progress-download]: https://github.com/charmbracelet/bubbletea/blob/master/examples/progress-download/main.go
[basic]: https://github.com/charmbracelet/bubbletea/blob/master/examples/composable-views/main.go
[glow]: https://github.com/charmbracelet/glow/blob/f0734709f0be19a34e648caaf63340938a50caa2/ui/ui.go#L434
[spinner]: https://github.com/charmbracelet/bubbles/blob/master/spinner/spinner.go#L142-L168
