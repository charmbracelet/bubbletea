# Common Patterns in Bubble Tea

So you've started building your app, but now you're not sure if you're doing
things the "right way". 

Well thankfully, we have some common patterns that we see when building Bubble
Tea applications that should help to simplify your decision-making.

## I only want the model that triggered the message to update

To figure out whether a component should process the message or not, simply
include an ID in the message. The ID will match the ID field of your child
model and can be handled in that child model's `Update`.

We use this pattern in our [spinner example](https://github.com/charmbracelet/bubbles/blob/master/spinner/spinner.go)

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
