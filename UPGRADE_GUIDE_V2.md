# Bubble Tea v2 Upgrade Guide

This guide covers everything you need to change when upgrading from Bubble Tea v1 to v2. If the compiler yells at you, this is the place to look. For a tour of all the exciting new features, check out the [What's New](WHATS_NEW_V2.md) doc.

> [!NOTE]
> We don't take API changes lightly and strive to make the upgrade process as simple as possible. If something feels way off, let us know.

## Migration Checklist

Here's the short version — a checklist you can follow top to bottom. Each item links to the relevant section below.

- [ ] [Update import paths](#import-paths)
- [ ] [Change `View() string` to `View() tea.View`](#view-returns-a-teaview-now)
- [ ] [Replace `tea.KeyMsg` with `tea.KeyPressMsg`](#key-messages)
- [ ] [Update key fields: `msg.Type` / `msg.Runes` / `msg.Alt`](#key-messages)
- [ ] [Replace `case " ":` with `case "space":`](#key-messages)
- [ ] [Update mouse message usage](#mouse-messages)
- [ ] [Rename mouse button constants](#mouse-messages)
- [ ] [Remove old program options → use View fields](#removed-program-options)
- [ ] [Remove imperative commands → use View fields](#removed-commands)
- [ ] [Remove old program methods](#removed-program-methods)
- [ ] [Rename `tea.WindowSize()` → `tea.RequestWindowSize`](#renamed-apis)
- [ ] [Replace `tea.Sequentially(...)` → `tea.Sequence(...)`](#renamed-apis)

## Import Paths

The module path changed to a vanity domain. Lip Gloss moved too.

```go
// Before
import tea "github.com/charmbracelet/bubbletea"
import "github.com/charmbracelet/lipgloss"

// After
import tea "charm.land/bubbletea/v2"
import "charm.land/lipgloss/v2"
```

## The Big Idea: Declarative Views

The single biggest change in v2 is the shift from **imperative commands** to **declarative View fields**. In v1, you'd use program options like `tea.WithAltScreen()` and commands like `tea.EnterAltScreen` to toggle terminal features on and off. In v2, you just set fields on the `tea.View` struct in your `View()` method and Bubble Tea handles the rest.

This means: no more startup option flags, no more toggle commands, no more fighting over state. Just declare what you want and Bubble Tea will make it so.

```go
// v1: imperative — scattered across NewProgram, Init, and Update
p := tea.NewProgram(model{}, tea.WithAltScreen(), tea.WithMouseCellMotion())

// v2: declarative — everything lives in View()
func (m model) View() tea.View {
    v := tea.NewView("Hello!")
    v.AltScreen = true
    v.MouseMode = tea.MouseModeCellMotion
    return v
}
```

Keep this in mind as you go through the rest of the guide — most of the "removed" things simply moved into View fields.

## View Returns a `tea.View` Now

The `View()` method no longer returns a `string`. It returns a `tea.View` struct.

```go
// Before:
func (m model) View() string {
    return "Hello, world!"
}

// After:
func (m model) View() tea.View {
    return tea.NewView("Hello, world!")
}
```

You can also use the longer form if you need to set additional fields:

```go
func (m model) View() tea.View {
    var v tea.View
    v.SetContent("Hello, world!")
    v.AltScreen = true
    return v
}
```

The `tea.View` struct has fields for everything that used to be controlled by options and commands:

| View Field | What It Does |
|---|---|
| `Content` | The rendered string (set via `SetContent()` or `NewView()`) |
| `AltScreen` | Enter/exit the alternate screen buffer |
| `MouseMode` | `MouseModeNone`, `MouseModeCellMotion`, or `MouseModeAllMotion` |
| `ReportFocus` | Enable focus/blur event reporting |
| `DisableBracketedPasteMode` | Disable bracketed paste |
| `WindowTitle` | Set the terminal window title |
| `Cursor` | Control cursor position, shape, color, and blink |
| `ForegroundColor` | Set the terminal foreground color |
| `BackgroundColor` | Set the terminal background color |
| `ProgressBar` | Show a native terminal progress bar |
| `KeyboardEnhancements` | Request keyboard enhancement features |
| `OnMouse` | Intercept mouse messages based on view content |

## Key Messages

Key messages got a major overhaul. Here's the quick rundown:

### `tea.KeyMsg` is now an interface

In v1, `tea.KeyMsg` was a struct you'd match on for key presses. In v2, it's an **interface** that covers both key presses and releases. For most code, you want `tea.KeyPressMsg`:

```go
// Before:
case tea.KeyMsg:
    switch msg.String() {
    case "q":
        return m, tea.Quit
    }

// After:
case tea.KeyPressMsg:
    switch msg.String() {
    case "q":
        return m, tea.Quit
    }
```

If you want to handle both presses and releases, use `tea.KeyMsg` and type-switch inside:

```go
case tea.KeyMsg:
    switch key := msg.(type) {
    case tea.KeyPressMsg:
        // key press
    case tea.KeyReleaseMsg:
        // key release
    }
```

### Key fields changed

| v1 | v2 | Notes |
|---|---|---|
| `msg.Type` | `msg.Code` | A `rune` — can be `tea.KeyEnter`, `'a'`, etc. |
| `msg.Runes` | `msg.Text` | Now a `string`, not `[]rune` |
| `msg.Alt` | `msg.Mod` | `msg.Mod.Contains(tea.ModAlt)` for alt, etc. |
| `tea.KeyRune` | — | Check `len(msg.Text) > 0` instead |
| `tea.KeyCtrlC` | — | Use `msg.String() == "ctrl+c"` or check `msg.Code` + `msg.Mod` |

### Space bar changed

Space bar now returns `"space"` instead of `" "` when using `msg.String()`:

```go
// Before:
case " ":

// After:
case "space":
```

`key.Code` is still `' '` and `key.Text` is still `" "`, but `String()` returns `"space"`.

### Ctrl+key matching

```go
// Before:
case tea.KeyCtrlC:
    // ctrl+c

// After (option A — string matching):
case tea.KeyPressMsg:
    switch msg.String() {
    case "ctrl+c":
        // ctrl+c
    }

// After (option B — field matching):
case tea.KeyPressMsg:
    if msg.Code == 'c' && msg.Mod == tea.ModCtrl {
        // ctrl+c
    }
```

### New Key fields

These are new in v2 and don't have v1 equivalents:

- **`key.ShiftedCode`** — the shifted key code (e.g., `'B'` when pressing shift+b)
- **`key.BaseCode`** — the key on a US PC-101 layout (handy for international keyboards)
- **`key.IsRepeat`** — whether the key is auto-repeating (Kitty protocol / Windows Console only)
- **`key.Keystroke()`** — like `String()` but always includes modifier info

## Paste Messages

Paste events no longer come in as `tea.KeyMsg` with a `Paste` flag. They're now their own message types:

```go
// Before:
case tea.KeyMsg:
    if msg.Paste {
        m.text += string(msg.Runes)
    }

// After:
case tea.PasteMsg:
    m.text += msg.Content
case tea.PasteStartMsg:
    // paste started
case tea.PasteEndMsg:
    // paste ended
```

## Mouse Messages

### `tea.MouseMsg` is now an interface

In v1, `tea.MouseMsg` was a struct with `X`, `Y`, `Button`, etc. In v2, it's an **interface**. You get the coordinates by calling `msg.Mouse()`:

```go
// Before:
case tea.MouseMsg:
    x, y := msg.X, msg.Y

// After:
case tea.MouseMsg:
    mouse := msg.Mouse()
    x, y := mouse.X, mouse.Y
```

### Mouse events are split by type

Instead of checking `msg.Action`, match on specific message types:

```go
// Before:
case tea.MouseMsg:
    if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
        // left click
    }

// After:
case tea.MouseClickMsg:
    if msg.Button == tea.MouseLeft {
        // left click
    }
case tea.MouseReleaseMsg:
    // release
case tea.MouseWheelMsg:
    // scroll
case tea.MouseMotionMsg:
    // movement
```

### Button constants renamed

| v1 | v2 |
|---|---|
| `tea.MouseButtonLeft` | `tea.MouseLeft` |
| `tea.MouseButtonRight` | `tea.MouseRight` |
| `tea.MouseButtonMiddle` | `tea.MouseMiddle` |
| `tea.MouseButtonWheelUp` | `tea.MouseWheelUp` |
| `tea.MouseButtonWheelDown` | `tea.MouseWheelDown` |
| `tea.MouseButtonWheelLeft` | `tea.MouseWheelLeft` |
| `tea.MouseButtonWheelRight` | `tea.MouseWheelRight` |

### `tea.MouseEvent` → `tea.Mouse`

The `MouseEvent` struct is gone. The new `Mouse` struct has `X`, `Y`, `Button`, and `Mod` fields.

### Mouse mode is now a View field

```go
// Before:
p := tea.NewProgram(model{}, tea.WithMouseCellMotion())

// After:
func (m model) View() tea.View {
    v := tea.NewView("...")
    v.MouseMode = tea.MouseModeCellMotion
    return v
}
```

## Removed Program Options

These options no longer exist. They all moved to View fields.

| Removed Option | Do This Instead |
|---|---|
| `tea.WithAltScreen()` | `view.AltScreen = true` |
| `tea.WithMouseCellMotion()` | `view.MouseMode = tea.MouseModeCellMotion` |
| `tea.WithMouseAllMotion()` | `view.MouseMode = tea.MouseModeAllMotion` |
| `tea.WithReportFocus()` | `view.ReportFocus = true` |
| `tea.WithoutBracketedPaste()` | `view.DisableBracketedPasteMode = true` |
| `tea.WithInputTTY()` | Just remove it — v2 always opens the TTY for input automatically |
| `tea.WithANSICompressor()` | Just remove it — the new renderer handles optimization automatically |

## Removed Commands

These commands no longer exist. Set the corresponding View field instead.

| Removed Command | Do This Instead |
|---|---|
| `tea.EnterAltScreen` | `view.AltScreen = true` |
| `tea.ExitAltScreen` | `view.AltScreen = false` |
| `tea.EnableMouseCellMotion` | `view.MouseMode = tea.MouseModeCellMotion` |
| `tea.EnableMouseAllMotion` | `view.MouseMode = tea.MouseModeAllMotion` |
| `tea.DisableMouse` | `view.MouseMode = tea.MouseModeNone` |
| `tea.HideCursor` | `view.Cursor = nil` |
| `tea.ShowCursor` | `view.Cursor = &tea.Cursor{...}` or `tea.NewCursor(x, y)` |
| `tea.EnableBracketedPaste` | `view.DisableBracketedPasteMode = false` |
| `tea.DisableBracketedPaste` | `view.DisableBracketedPasteMode = true` |
| `tea.EnableReportFocus` | `view.ReportFocus = true` |
| `tea.DisableReportFocus` | `view.ReportFocus = false` |
| `tea.SetWindowTitle("...")` | `view.WindowTitle = "..."` |

## Removed Program Methods

These methods on `*Program` are gone.

| Removed Method | Do This Instead |
|---|---|
| `p.Start()` | `p.Run()` |
| `p.StartReturningModel()` | `p.Run()` |
| `p.EnterAltScreen()` | `view.AltScreen = true` in `View()` |
| `p.ExitAltScreen()` | `view.AltScreen = false` in `View()` |
| `p.EnableMouseCellMotion()` | `view.MouseMode` in `View()` |
| `p.DisableMouseCellMotion()` | `view.MouseMode = tea.MouseModeNone` in `View()` |
| `p.EnableMouseAllMotion()` | `view.MouseMode` in `View()` |
| `p.DisableMouseAllMotion()` | `view.MouseMode = tea.MouseModeNone` in `View()` |
| `p.SetWindowTitle(...)` | `view.WindowTitle` in `View()` |

## Renamed APIs

| v1 | v2 | Notes |
|---|---|---|
| `tea.Sequentially(...)` | `tea.Sequence(...)` | `Sequentially` was already deprecated in v1 |
| `tea.WindowSize()` | `tea.RequestWindowSize` | Now returns `Msg` directly, not a `Cmd` |

## New Program Options

These are new in v2:

| Option | What It Does |
|---|---|
| `tea.WithColorProfile(p)` | Force a specific color profile (great for testing) |
| `tea.WithWindowSize(w, h)` | Set initial terminal size (great for testing) |

## Complete Before & After

Here's a minimal but complete program showing the most common migration patterns side by side.

**v1:**

```go
package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    count int
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case " ":
            m.count++
        }
    case tea.MouseMsg:
        if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
            m.count++
        }
    }
    return m, nil
}

func (m model) View() string {
    return fmt.Sprintf("Count: %d\n\nSpace or click to increment. q to quit.\n", m.count)
}

func main() {
    p := tea.NewProgram(model{}, tea.WithAltScreen(), tea.WithMouseCellMotion())
    if _, err := p.Run(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

**v2:**

```go
package main

import (
    "fmt"
    "os"

    tea "charm.land/bubbletea/v2"
)

type model struct {
    count int
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "space":
            m.count++
        }
    case tea.MouseClickMsg:
        if msg.Button == tea.MouseLeft {
            m.count++
        }
    }
    return m, nil
}

func (m model) View() tea.View {
    v := tea.NewView(fmt.Sprintf("Count: %d\n\nSpace or click to increment. q to quit.\n", m.count))
    v.AltScreen = true
    v.MouseMode = tea.MouseModeCellMotion
    return v
}

func main() {
    p := tea.NewProgram(model{})
    if _, err := p.Run(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

Notice how the `NewProgram` call got simpler? All the terminal feature flags moved into `View()` where they belong.

## Quick Reference

A flat old → new lookup table. Handy for search-and-replace and LLM-assisted migration.

### Import Paths

| v1 | v2 |
|---|---|
| `github.com/charmbracelet/bubbletea` | `charm.land/bubbletea/v2` |
| `github.com/charmbracelet/lipgloss` | `charm.land/lipgloss/v2` |

### Model Interface

| v1 | v2 |
|---|---|
| `View() string` | `View() tea.View` |

### Key Events

| v1 | v2 |
|---|---|
| `tea.KeyMsg` (struct) | `tea.KeyPressMsg` for presses, `tea.KeyMsg` (interface) for both |
| `msg.Type` | `msg.Code` |
| `msg.Runes` | `msg.Text` (string, not `[]rune`) |
| `msg.Alt` | `msg.Mod.Contains(tea.ModAlt)` |
| `tea.KeyRune` | check `len(msg.Text) > 0` |
| `tea.KeyCtrlC` | `msg.Code == 'c' && msg.Mod == tea.ModCtrl` or `msg.String() == "ctrl+c"` |
| `case " ":` (space) | `case "space":` |

### Mouse Events

| v1 | v2 |
|---|---|
| `tea.MouseMsg` (struct) | `tea.MouseMsg` (interface) — call `.Mouse()` for the data |
| `tea.MouseEvent` | `tea.Mouse` |
| `tea.MouseButtonLeft` | `tea.MouseLeft` |
| `tea.MouseButtonRight` | `tea.MouseRight` |
| `tea.MouseButtonMiddle` | `tea.MouseMiddle` |
| `tea.MouseButtonWheelUp` | `tea.MouseWheelUp` |
| `tea.MouseButtonWheelDown` | `tea.MouseWheelDown` |
| `msg.X`, `msg.Y` (direct) | `msg.Mouse().X`, `msg.Mouse().Y` |

### Options → View Fields

| v1 Option | v2 View Field |
|---|---|
| `tea.WithAltScreen()` | `view.AltScreen = true` |
| `tea.WithMouseCellMotion()` | `view.MouseMode = tea.MouseModeCellMotion` |
| `tea.WithMouseAllMotion()` | `view.MouseMode = tea.MouseModeAllMotion` |
| `tea.WithReportFocus()` | `view.ReportFocus = true` |
| `tea.WithoutBracketedPaste()` | `view.DisableBracketedPasteMode = true` |

### Commands → View Fields

| v1 Command | v2 View Field |
|---|---|
| `tea.EnterAltScreen` / `tea.ExitAltScreen` | `view.AltScreen = true/false` |
| `tea.EnableMouseCellMotion` | `view.MouseMode = tea.MouseModeCellMotion` |
| `tea.EnableMouseAllMotion` | `view.MouseMode = tea.MouseModeAllMotion` |
| `tea.DisableMouse` | `view.MouseMode = tea.MouseModeNone` |
| `tea.HideCursor` / `tea.ShowCursor` | `view.Cursor = nil` / `view.Cursor = &tea.Cursor{...}` |
| `tea.EnableBracketedPaste` / `tea.DisableBracketedPaste` | `view.DisableBracketedPasteMode = false/true` |
| `tea.EnableReportFocus` / `tea.DisableReportFocus` | `view.ReportFocus = true/false` |
| `tea.SetWindowTitle("...")` | `view.WindowTitle = "..."` |

### Removed Options (No Replacement Needed)

| v1 Option | What Happened |
|---|---|
| `tea.WithInputTTY()` | v2 always opens the TTY for input automatically |
| `tea.WithANSICompressor()` | The new renderer handles optimization automatically |

### Removed Program Methods

| v1 Method | v2 Replacement |
|---|---|
| `p.Start()` | `p.Run()` |
| `p.StartReturningModel()` | `p.Run()` |
| `p.EnterAltScreen()` | `view.AltScreen = true` in `View()` |
| `p.ExitAltScreen()` | `view.AltScreen = false` in `View()` |
| `p.EnableMouseCellMotion()` | `view.MouseMode` in `View()` |
| `p.DisableMouseCellMotion()` | `view.MouseMode = tea.MouseModeNone` in `View()` |
| `p.EnableMouseAllMotion()` | `view.MouseMode` in `View()` |
| `p.DisableMouseAllMotion()` | `view.MouseMode = tea.MouseModeNone` in `View()` |
| `p.SetWindowTitle(...)` | `view.WindowTitle` in `View()` |

### Other Renames

| v1 | v2 |
|---|---|
| `tea.Sequentially(...)` | `tea.Sequence(...)` |
| `tea.WindowSize()` | `tea.RequestWindowSize` (now returns `Msg`, not `Cmd`) |

### New Program Options

| Option | Description |
|---|---|
| `tea.WithColorProfile(p)` | Force a specific color profile |
| `tea.WithWindowSize(w, h)` | Set initial window size (great for testing) |

## Feedback

Have thoughts on the v2 upgrade? We'd _love_ to hear about it. Let us know on…

- [Discord](https://charm.land/chat)
- [Matrix](https://charm.land/matrix)
- [Email](mailto:vt100@charm.land)

---

Part of [Charm](https://charm.land).

<a href="https://charm.land/"><img alt="The Charm logo" src="https://stuff.charm.land/charm-badge.jpg" width="400"></a>

Charm热爱开源 • Charm loves open source • نحنُ نحب المصادر المفتوحة
