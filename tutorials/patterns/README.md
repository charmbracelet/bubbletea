# Common Patterns in Bubble Tea

So you've started building your app, but now you're not sure if you're doing
things the "right way". 

Well thankfully, we have some common patterns that we see when building Bubble
Tea applications that should help to simplify your decision-making.

## Managing multiple components in one model

If you have a composite view, then you have multiple components on one screen
that you want to be able to switch between. To handle this in Bubble Tea you'll
want your parent component to house a `state` field that dictates which element
on the screen is focused and receiving key presses.

You can see a [basic example][basic] of this on our GitHub.

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

## I only want the model that triggered the message to update

To figure out whether a component should process the message or not, simply
include an ID in the message. The ID will match the ID field of your child
model and can be handled in that child model's `Update`.

We use this pattern in our [spinner example][spinner]

These spots in particular:
https://github.com/charmbracelet/bubbles/blob/master/spinner/spinner.go#L145-L149

```go
		// If an ID is set, and the ID doesn't belong to this spinner, reject
		// the message.
		if msg.ID > 0 && msg.ID != m.id {
			return m, nil
		}
```

https://github.com/charmbracelet/bubbles/blob/master/spinner/spinner.go#L164
https://github.com/charmbracelet/bubbles/blob/master/spinner/spinner.go#L195-L203l

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

[basic]: https://github.com/charmbracelet/bubbletea/blob/master/examples/composable-views/main.go
[glow]: https://github.com/charmbracelet/glow/blob/f0734709f0be19a34e648caaf63340938a50caa2/ui/ui.go#L434
[spinner]: https://github.com/charmbracelet/bubbles/blob/master/spinner/spinner.go
