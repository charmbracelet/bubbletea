# Frequently Asked Questions

These are some of the most commonly asked questions regarding the `list` bubble.

## Adding Custom Items

There are a few things you need to do to create custom items. First off, they
need to implement the `list.Item` and `list.DefaultItem` interfaces.

```go
// Item is an item that appears in the list.
type Item interface {
	// FilterValue is the value we use when filtering against this item when
	// we're filtering the list.
	FilterValue() string
}
```

```go
// DefaultItem describes an items designed to work with DefaultDelegate.
type DefaultItem interface {
	Item
	Title() string
	Description() string
}
```

You can see a working example in our [Kancli][kancli] project built
explicitly for a tutorial on lists and composite views in Bubble Tea. 

[VIDEO](https://youtu.be/ZA93qgdLUzM)

## Customizing Styles

Rendering (and behavior) for list items is done via the
[`ItemDelegate`][itemDelegate]
interface. It can be a little confusing at first, but it allows the list to be
very flexible and powerful.

If you just want to alter the default style you could do something like:

```go
import "github.com/rprtr258/bubbletea/bubbles/list"

// Create a new default delegate
d := list.NewDefaultDelegate()

// Change colors
c := lipgloss.Color("#6f03fc")
d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(c).BorderLeftForeground(c)
d.Styles.SelectedDesc = d.Styles.SelectedTitle.Copy() // reuse the title style here

// Initailize the list model with our delegate
width, height := 80, 40
l := list.New(listItems, d, width, height)

// You can also change the delegate on the fly
l.SetDelegate(d)
```

This code would replace [this line][replacedLine] in the [`list-default`
example][listDefault].

For full control over the way list items are rendered you can also define your
own `ItemDelegate` too ([example][customDelegate]).


[kancli]: https://github.com/charmbracelet/kancli/blob/main/main.go#L45
[itemDelegate]: https://pkg.go.dev/github.com/rprtr258/bubbletea/bubbles@v0.10.2/list#ItemDelegate
[replacedLine]: https://github.com/rprtr258/bubbletea/blob/master/examples/list-default/main.go#L77
[listDefault]: https://github.com/rprtr258/bubbletea/tree/master/examples/list-default
[customDelegate]: https://github.com/rprtr258/bubbletea/blob/a6f46172ec4436991b90c2270253b2d212de7ef3/examples/list-simple/main.go#L28-L49
