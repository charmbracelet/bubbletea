Lip Gloss
=========

<p>
    <img src="https://stuff.charm.sh/lipgloss/lipgloss-header-github.png" width="340" alt="Lip Gloss Title Treatment"><br>
    <a href="https://github.com/rprtr258/bubbletea/lipgloss/releases"><img src="https://img.shields.io/github/release/rprtr258/bubbletea/lipgloss.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/github.com/rprtr258/bubbletea/lipgloss?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc"></a>
    <a href="https://github.com/rprtr258/bubbletea/lipgloss/actions"><img src="https://github.com/rprtr258/bubbletea/lipgloss/workflows/build/badge.svg" alt="Build Status"></a>
</p>

Style definitions for nice terminal layouts. Built with TUIs in mind.

![Lip Gloss example](https://stuff.charm.sh/lipgloss/lipgloss-example.png)

Lip Gloss takes an expressive, declarative approach to terminal rendering.
Users familiar with CSS will feel at home with Lip Gloss.

```go

import "github.com/rprtr258/bubbletea/lipgloss"

var style = lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#FAFAFA")).
    Background(lipgloss.Color("#7D56F4")).
    PaddingTop(2).
    PaddingLeft(4).
    Width(22)

fmt.Println(style.Render("Hello, kitty"))
```

## Colors

Lip Gloss supports the following color profiles:

### ANSI 16 colors (4-bit)

```go
lipgloss.Color("5")  // magenta
lipgloss.Color("9")  // red
lipgloss.Color("12") // light blue
```

### ANSI 256 Colors (8-bit)

```go
lipgloss.Color("86")  // aqua
lipgloss.Color("201") // hot pink
lipgloss.Color("202") // orange
```

### True Color (16,777,216 colors; 24-bit)

```go
lipgloss.Color("#0000FF") // good ol' 100% blue
lipgloss.Color("#04B575") // a green
lipgloss.Color("#3C3C3C") // a dark gray
```

...as well as a 1-bit ASCII profile, which is black and white only.

The terminal's color profile will be automatically detected, and colors outside
the gamut of the current palette will be automatically coerced to their closest
available value.


### Adaptive Colors

You can also specify color options for light and dark backgrounds:

```go
lipgloss.AdaptiveColor{Light: "236", Dark: "248"}
```

The terminal's background color will automatically be detected and the
appropriate color will be chosen at runtime.

### Complete Colors

CompleteColor specifies exact values for truecolor, ANSI256, and ANSI color
profiles.

```go
lipgloss.CompleteColor{True: "#0000FF", ANSI256: "86", ANSI: "5"}
```

Automatic color degradation will not be performed in this case and it will be
based on the color specified.

### Complete Adaptive Colors

You can use CompleteColor with AdaptiveColor to specify the exact values for
light and dark backgrounds without automatic color degradation.

```go
lipgloss.CompleteAdaptiveColor{
    Light: CompleteColor{TrueColor: "#d7ffae", ANSI256: "193", ANSI: "11"},
    Dark:  CompleteColor{TrueColor: "#d75fee", ANSI256: "163", ANSI: "5"},
}
```

## Inline Formatting

Lip Gloss supports the usual ANSI text formatting options:

```go
var style = lipgloss.NewStyle().
    Bold(true).
    Italic(true).
    Faint(true).
    Blink(true).
    Strikethrough(true).
    Underline(true).
    Reverse(true)
```


## Block-Level Formatting

Lip Gloss also supports rules for block-level formatting:

```go
// Padding
var style = lipgloss.NewStyle().
    PaddingTop(2).
    PaddingRight(4).
    PaddingBottom(2).
    PaddingLeft(4)

// Margins
var style = lipgloss.NewStyle().
    MarginTop(2).
    MarginRight(4).
    MarginBottom(2).
    MarginLeft(4)
```

There is also shorthand syntax for margins and padding, which follows the same
format as CSS:

```go
// 2 cells on all sides
lipgloss.NewStyle().Padding(2)

// 2 cells on the top and bottom, 4 cells on the left and right
lipgloss.NewStyle().Margin(2, 4)

// 1 cell on the top, 4 cells on the sides, 2 cells on the bottom
lipgloss.NewStyle().Padding(1, 4, 2)

// Clockwise, starting from the top: 2 cells on the top, 4 on the right, 3 on
// the bottom, and 1 on the left
lipgloss.NewStyle().Margin(2, 4, 3, 1)
```


## Aligning Text

You can align paragraphs of text to the left, right, or center.

```go
var style = lipgloss.NewStyle().
    Width(24).
    Align(lipgloss.Left).  // align it left
    Align(lipgloss.Right). // no wait, align it right
    Align(lipgloss.Center) // just kidding, align it in the center
```


## Width and Height

Setting a minimum width and height is simple and straightforward.

```go
var style = lipgloss.NewStyle().
    SetString("What’s for lunch?").
    Width(24).
    Height(32).
    Foreground(lipgloss.Color("63"))
```


## Borders

Adding borders is easy:

```go
// Add a purple, rectangular border
var style = lipgloss.NewStyle().
    BorderStyle(lipgloss.NormalBorder()).
    BorderForeground(lipgloss.Color("63"))

// Set a rounded, yellow-on-purple border to the top and left
var anotherStyle = lipgloss.NewStyle().
    BorderStyle(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("228")).
    BorderBackground(lipgloss.Color("63")).
    BorderTop(true).
    BorderLeft(true)

// Make your own border
var myCuteBorder = lipgloss.Border{
    Top:         "._.:*:",
    Bottom:      "._.:*:",
    Left:        "|*",
    Right:       "|*",
    TopLeft:     "*",
    TopRight:    "*",
    BottomLeft:  "*",
    BottomRight: "*",
}
```

There are also shorthand functions for defining borders, which follow a similar
pattern to the margin and padding shorthand functions.

```go
// Add a thick border to the top and bottom
lipgloss.NewStyle().
    Border(lipgloss.ThickBorder(), true, false)

// Add a thick border to the right and bottom sides. Rules are set clockwise
// from top.
lipgloss.NewStyle().
    Border(lipgloss.DoubleBorder(), true, false, false, true)
```

For more on borders see [the docs][docs].


## Copying Styles

Just use `Copy()`:

```go
var style = lipgloss.NewStyle().Foreground(lipgloss.Color("219"))

var wildStyle = style.Copy().Blink(true)
```

`Copy()` performs a copy on the underlying data structure ensuring that you get
a true, dereferenced copy of a style. Without copying, it's possible to mutate
styles.


## Inheritance

Styles can inherit rules from other styles. When inheriting, only unset rules
on the receiver are inherited.

```go
var styleA = lipgloss.NewStyle().
    Foreground(lipgloss.Color("229")).
    Background(lipgloss.Color("63"))

// Only the background color will be inherited here, because the foreground
// color will have been already set:
var styleB = lipgloss.NewStyle().
    Foreground(lipgloss.Color("201")).
    Inherit(styleA)
```


## Unsetting Rules

All rules can be unset:

```go
var style = lipgloss.NewStyle().
    Bold(true).                        // make it bold
    UnsetBold().                       // jk don't make it bold
    Background(lipgloss.Color("227")). // yellow background
    UnsetBackground()                  // never mind
```

When a rule is unset, it won't be inherited or copied.


## Enforcing Rules

Sometimes, such as when developing a component, you want to make sure style
definitions respect their intended purpose in the UI. This is where `Inline`
and `MaxWidth`, and `MaxHeight` come in:

```go
// Force rendering onto a single line, ignoring margins, padding, and borders.
someStyle.Inline(true).Render("yadda yadda")

// Also limit rendering to five cells
someStyle.Inline(true).MaxWidth(5).Render("yadda yadda")

// Limit rendering to a 5x5 cell block
someStyle.MaxWidth(5).MaxHeight(5).Render("yadda yadda")
```

## Rendering

Generally, you just call the `Render(string...)` method on a `lipgloss.Style`:

```go
style := lipgloss.NewStyle().Bold(true).SetString("Hello,")
fmt.Println(style.Render("kitty.")) // Hello, kitty.
fmt.Println(style.Render("puppy.")) // Hello, puppy.
```

But you could also use the Stringer interface:

```go
var style = lipgloss.NewStyle().SetString("你好，猫咪。").Bold(true)
fmt.Println(style) // 你好，猫咪。
```

### Custom Renderers

Custom renderers allow you to render to a specific outputs. This is
particularly important when you want to render to different outputs and
correctly detect the color profile and dark background status for each, such as
in a server-client situation.

```go
func myLittleHandler(sess ssh.Session) {
    // Create a renderer for the client.
    renderer := lipgloss.NewRenderer(sess)

    // Create a new style on the renderer.
    style := renderer.NewStyle().Background(lipgloss.AdaptiveColor{Light: "63", Dark: "228"})

    // Render. The color profile and dark background state will be correctly detected.
    io.WriteString(sess, style.Render("Heyyyyyyy"))
}
```

For an example on using a custom renderer over SSH with [Wish][wish] see the
[SSH example][ssh-example].

## Utilities

In addition to pure styling, Lip Gloss also ships with some utilities to help
assemble your layouts.


### Joining Paragraphs

Horizontally and vertically joining paragraphs is a cinch.

```go
// Horizontally join three paragraphs along their bottom edges
lipgloss.JoinHorizontal(lipgloss.Bottom, paragraphA, paragraphB, paragraphC)

// Vertically join two paragraphs along their center axes
lipgloss.JoinVertical(lipgloss.Center, paragraphA, paragraphB)

// Horizontally join three paragraphs, with the shorter ones aligning 20%
// from the top of the tallest
lipgloss.JoinHorizontal(0.2, paragraphA, paragraphB, paragraphC)
```


### Measuring Width and Height

Sometimes you’ll want to know the width and height of text blocks when building
your layouts.

```go
// Render a block of text.
var style = lipgloss.NewStyle().
    Width(40).
    Padding(2)
var block string = style.Render(someLongString)

// Get the actual, physical dimensions of the text block.
width := lipgloss.Width(block)
height := lipgloss.Height(block)

// Here's a shorthand function.
w, h := lipgloss.Size(block)
```


### Placing Text in Whitespace

Sometimes you’ll simply want to place a block of text in whitespace.

```go
// Center a paragraph horizontally in a space 80 cells wide. The height of
// the block returned will be as tall as the input paragraph.
block := lipgloss.PlaceHorizontal(80, lipgloss.Center, fancyStyledParagraph)

// Place a paragraph at the bottom of a space 30 cells tall. The width of
// the text block returned will be as wide as the input paragraph.
block := lipgloss.PlaceVertical(30, lipgloss.Bottom, fancyStyledParagraph)

// Place a paragraph in the bottom right corner of a 30x80 cell space.
block := lipgloss.Place(30, 80, lipgloss.Right, lipgloss.Bottom, fancyStyledParagraph)
```

You can also style the whitespace. For details, see [the docs][docs].


***

## FAQ

<details>
<summary>
Why are things misaligning? Why are borders at the wrong widths?
</summary>
<p>This is most likely due to your locale and encoding, particularly with
regard to Chinese, Japanese, and Korean (for example, <code>zh_CN.UTF-8</code>
or <code>ja_JP.UTF-8</code>). The most direct way to fix this is to set
<code>RUNEWIDTH_EASTASIAN=0</code> in your environment.</p>

<p>For details see <a href="https://github.com/rprtr258/bubbletea/lipgloss/issues/40">https://github.com/rprtr258/bubbletea/lipgloss/issues/40.</a></p>
</details>

<details>
<summary>
Why isn't Lip Gloss displaying colors?
</summary>
<p>Lip Gloss automatically degrades colors to the best available option in the
given terminal, and if output's not a TTY it will remove color output entirely.
This is common when running tests, CI, or when piping output elsewhere.</p>

<p>If necessary, you can force a color profile in your tests with
<a href="https://pkg.go.dev/github.com/rprtr258/bubbletea/lipgloss#SetColorProfile"><code>SetColorProfile</code></a>.</p>

```go
import (
    "github.com/rprtr258/bubbletea/lipgloss"
    "github.com/muesli/termenv"
)

lipgloss.SetColorProfile(termenv.TrueColor)
```

*Note:* this option limits the flexibility of your application and can cause
ANSI escape codes to be output in cases where that might not be desired. Take
careful note of your use case and environment before choosing to force a color
profile.
</details>

## What about [Bubble Tea][tea]?

Lip Gloss doesn’t replace Bubble Tea. Rather, it is an excellent Bubble Tea
companion. It was designed to make assembling terminal user interface views as
simple and fun as possible so that you can focus on building your application
instead of concerning yourself with low-level layout details.

In simple terms, you can use Lip Gloss to help build your Bubble Tea views.

[tea]: https://github.com/charmbracelet/tea


## Under the Hood

Lip Gloss is built on the excellent [Termenv][termenv] and [Reflow][reflow]
libraries which deal with color and ANSI-aware text operations, respectively.
For many use cases Termenv and Reflow will be sufficient for your needs.

[termenv]: https://github.com/muesli/termenv
[reflow]: https://github.com/muesli/reflow


## Rendering Markdown

For a more document-centric rendering solution with support for things like
lists, tables, and syntax-highlighted code have a look at [Glamour][glamour],
the stylesheet-based Markdown renderer.

[glamour]: https://github.com/rprtr258/bubbletea/glamour


## Feedback

We’d love to hear your thoughts on this project. Feel free to drop us a note!

* [Twitter](https://twitter.com/charmcli)
* [The Fediverse](https://mastodon.social/@charmcli)
* [Discord](https://charm.sh/chat)

## License

[MIT](https://github.com/rprtr258/bubbletea/lipgloss/raw/master/LICENSE)

***

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400"></a>

Charm热爱开源 • Charm loves open source


[docs]: https://pkg.go.dev/github.com/rprtr258/bubbletea/lipgloss?tab=doc
[wish]: https://github.com/charmbracelet/wish
[ssh-example]: examples/ssh
