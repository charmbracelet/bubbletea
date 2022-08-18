Bubble Tea
==========

<p>
    <img src="https://stuff.charm.sh/bubbletea/bubbletea-github-header-simple.png" width="313" alt="Bubble Tea Title Treatment"><br>
    <a href="https://github.com/charmbracelet/bubbletea/releases"><img src="https://img.shields.io/github/release/charmbracelet/bubbletea.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/github.com/charmbracelet/bubbletea?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc"></a>
    <a href="https://github.com/charmbracelet/bubbletea/actions"><img src="https://github.com/charmbracelet/bubbletea/workflows/build/badge.svg" alt="Build Status"></a>
</p>

The fun, functional and stateful way to build terminal apps. A Go framework
based on [The Elm Architecture][elm]. Bubble Tea is well-suited for simple and
complex terminal applications, either inline, full-window, or a mix of both.

<p>
    <img src="https://stuff.charm.sh/bubbletea/bubbletea-example.gif?0" width="800" alt="Bubble Tea Example">
</p>

Bubble Tea is in use in production and includes a number of features and
performance optimizations we’ve added along the way. Among those is a standard
framerate-based renderer, a renderer for high-performance scrollable
regions which works alongside the main renderer, and mouse support.

## Getting Started

We recommend starting with the [basics tutorial][basics] followed by the
[commands tutorial][commands], both of which should give you a good
understanding of how things work.

There are a bunch of [examples][examples], too!

[basics]: https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics
[commands]: https://github.com/charmbracelet/bubbletea/tree/master/tutorials/commands
[documentation]: https://github.com/charmbracelet/bubbletea/tree/master/docs
[examples]: https://github.com/charmbracelet/bubbletea/tree/master/examples

## Components

For a bunch of basic user interface components check out [Bubbles][bubbles],
the official Bubble Tea component library.

<p>
    <a href="https://github.com/charmbracelet/bubbles"><img src="https://stuff.charm.sh/bubbles/bubbles-badge.png" width="174" alt="Bubbles Badge"></a>&nbsp;&nbsp;
    <a href="https://github.com/charmbracelet/bubbles"><img src="https://stuff.charm.sh/bubbles-examples/textinput.gif" width="400" alt="Text Input Example from Bubbles"></a>
</p>

## Debugging

### Debugging with Delve

Since Bubble Tea apps assume control of stdin and stdout, you’ll need to run
delve in headless mode and then connect to it:

```bash
# Start the debugger
$ dlv debug --headless .
API server listening at: 127.0.0.1:34241

# Connect to it from another terminal
$ dlv connect 127.0.0.1:34241
```

Note that the default port used will vary on your system and per run, so
actually watch out what address the first `dlv` run tells you to connect to.

### Logging Stuff

You can log to a debug file to print debug Bubble Tea applications. To do so,
include something like…

```go
if len(os.Getenv("DEBUG")) > 0 {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()
}
```

…before you start your Bubble Tea program. To see what’s printed in real time,
run `tail -f debug.log` while you run your program in another window.

## Libraries we use with Bubble Tea

* [Bubbles][bubbles]: Common Bubble Tea components such as text inputs, viewports, spinners and so on
* [Lip Gloss][lipgloss]: Style, format and layout tools for terminal applications
* [Harmonica][harmonica]: A spring animation library for smooth, natural motion
* [BubbleZone][bubblezone]: Easy mouse event tracking for Bubble Tea components
* [Termenv][termenv]: Advanced ANSI styling for terminal applications
* [Reflow][reflow]: Advanced ANSI-aware methods for working with text

[bubbles]: https://github.com/charmbracelet/bubbles
[lipgloss]: https://github.com/charmbracelet/lipgloss
[harmonica]: https://github.com/charmbracelet/harmonica
[bubblezone]: https://github.com/lrstanley/bubblezone
[termenv]: https://github.com/muesli/termenv
[reflow]: https://github.com/muesli/reflow

## Bubble Tea in the Wild

For some Bubble Tea programs in production, see:

* [AT CLI](https://github.com/daskycodes/at_cli): a utility for executing AT Commands via serial port connections
* [Aztify](https://github.com/Azure/aztfy): bring Microsoft Azure resources under Terraform
* [Canard](https://github.com/mrusme/canard): an RSS client
* [charm](https://github.com/charmbracelet/charm): the official Charm user account manager
* [circumflex](https://github.com/bensadeh/circumflex): read Hacker News in your terminal
* [clidle](https://github.com/ajeetdsouza/clidle): a Wordle clone for your terminal
* [container-canary](https://github.com/NVIDIA/container-canary): a container validator
* [dns53](https://github.com/purpleclay/dns53): dynamic DNS with Amazon Route53. Expose your EC2 quickly, securely and privately
* [fm](https://github.com/knipferrc/fm): a terminal-based file manager
* [flapioca](https://github.com/kbrgl/flapioca): Flappy Bird on the CLI!
* [fztea](https://github.com/jon4hz/fztea): connect to your Flipper's UI over serial or make it accessible via SSH 
* [fork-cleaner](https://github.com/caarlos0/fork-cleaner): cleans up old and inactive forks in your GitHub account
* [gambit](https://github.com/maaslalani/gambit): play chess in the terminal
* [gembro](https://git.sr.ht/~rafael/gembro): a mouse-driven Gemini browser
* [gh-b](https://github.com/joaom00/gh-b): GitHub CLI extension to easily manage your branches
* [gh-dash](https://www.github.com/dlvhdr/gh-dash): GitHub CLI extension to display a dashboard of PRs and issues
* [gitflow-toolkit](https://github.com/mritd/gitflow-toolkit): a GitFlow submission tool
* [Glow](https://github.com/charmbracelet/glow): a markdown reader, browser and online markdown stash
* [gocovsh](https://github.com/orlangure/gocovsh): explore Go coverage reports from the CLI
* [httpit](https://github.com/gonetx/httpit): a rapid http(s) benchmark tool
* [IDNT](https://github.com/r-darwish/idnt): batch software uninstaller
* [kboard](https://github.com/CamiloGarciaLaRotta/kboard): a typing game
* [mandelbrot-cli](https://github.com/MicheleFiladelfia/mandelbrot-cli): Multiplatform terminal mandelbrot set explorer
* [mergestat](https://github.com/mergestat/mergestat): run SQL queries on git repositories
* [mc](https://github.com/minio/mc): the official [MinIO](https://min.io) client
* [pathos](https://github.com/chip/pathos): pathos - CLI for editing a PATH env variable
* [portal][portal]: securely send transfer between computers
* [redis-viewer](https://github.com/SaltFishPr/redis-viewer): browse Redis databases
* [sku](https://github.com/fedeztk/sku): a simple TUI for playing sudoku inside the terminal
* [Slides](https://github.com/maaslalani/slides): a markdown-based presentation tool
* [Soft Serve](https://github.com/charmbracelet/soft-serve): a command-line-first Git server that runs a TUI over SSH
* [StormForge Optimize Controller](https://github.com/thestormforge/optimize-controller): a tool for experimenting with application configurations in Kubernetes
* [STTG](https://github.com/wille1101/sttg): teletext client for SVT, Sweden’s national public television station
* [sttr](https://github.com/abhimanyu003/sttr): run various text transformations
* [tasktimer](https://github.com/caarlos0/tasktimer): a dead-simple task timer
* [termdbms](https://github.com/mathaou/termdbms): a keyboard and mouse driven database browser
* [ticker](https://github.com/achannarasappa/ticker): a terminal stock watcher and stock position tracker
* [tran](https://github.com/abdfnx/tran): securely transfer stuff between computers (based on [portal][portal])
* [tz](https://github.com/oz/tz): an aid for scheduling across multiple time zones
* [ugm](https://github.com/ariasmn/ugm): a unix user and group browser
* [Typer](https://github.com/maaslalani/typer): a typing test
* [wishlist](https://github.com/charmbracelet/wishlist): an SSH directory
* [Noted](https://github.com/torbratsberg/noted): Note viewer and manager

[portal]: https://github.com/ZinoKader/portal

## Feedback

We'd love to hear your thoughts on this tutorial. Feel free to drop us a note!

* [Twitter](https://twitter.com/charmcli)
* [The Fediverse](https://mastodon.technology/@charm)
* [Slack](https://charm.sh/slack)

## Acknowledgments

Bubble Tea is based on the paradigms of [The Elm Architecture][elm] by Evan
Czaplicki et alia and the excellent [go-tea][gotea] by TJ Holowaychuk. It’s
inspired by the many great [_Zeichenorientierte Benutzerschnittstellen_][zb]
of days past.

[elm]: https://guide.elm-lang.org/architecture/
[gotea]: https://github.com/tj/go-tea
[zb]: https://de.wikipedia.org/wiki/Zeichenorientierte_Benutzerschnittstelle

## License

[MIT](https://github.com/charmbracelet/bubbletea/raw/master/LICENSE)

***

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400"></a>

Charm热爱开源 • Charm loves open source
