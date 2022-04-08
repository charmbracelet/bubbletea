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

To get started, see the tutorial below, the [examples][examples], the
[docs][docs] and some common [resources](#libraries-we-use-with-bubble-tea).

Note: If you're working with submodels, you don't need to worry about using commands to send data to your main model as the data flows downward. 
This means that the parent knows about the children. 

### So, when do things run?

**Init** - `Init()` is called when the program starts, its role is to fire off initial Commands  

**Update** - `Update()` runs when a command is triggered, this is any function that returns a `tea.Msg`  

**View** - `View()` is called automatically after `Update()` to redraw the program with the updated state.  

Definitely check out the Elm architecture resource above to learn more about how ELM works. Also, hop into our [Slack](https://charm.sh/slack) if you have any questions or want to be part of the community!

## By the way

Be sure to check out [Bubbles][bubbles], a library of common UI components for Bubble Tea.

<p>
    <a href="https://github.com/charmbracelet/bubbles"><img src="https://stuff.charm.sh/bubbles/bubbles-badge.png" width="174" alt="Bubbles Badge"></a>&nbsp;&nbsp;
    <a href="https://github.com/charmbracelet/bubbles"><img src="https://stuff.charm.sh/bubbles-examples/textinput.gif" width="400" alt="Text Input Example from Bubbles"></a>
</p>

* * *

## Getting Started

We recommend starting with our [basics tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics) followed by our [commands tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials/commands) to understand what you can do with `tea.Cmd` and how they're used in BubbleTea. 

Additionally, we have have [documentation](https://github.com/charmbracelet/bubbletea/tree/master/docs) that include frequently asked questions, issues, and more in-depth explanations on specific topics.
Don't forget about our [examples](https://github.com/charmbracelet/bubbletea/tree/master/examples).

## Debugging with Delve

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

## Libraries we use with Bubble Tea

* [Bubbles][bubbles]: Common Bubble Tea components such as text inputs, viewports, spinners and so on
* [Lip Gloss][lipgloss]: Style, format and layout tools for terminal applications
* [Harmonica][harmonica]: A spring animation library for smooth, natural motion
* [Termenv][termenv]: Advanced ANSI styling for terminal applications
* [Reflow][reflow]: Advanced ANSI-aware methods for working with text

[bubbles]: https://github.com/charmbracelet/bubbles
[lipgloss]: https://github.com/charmbracelet/lipgloss
[harmonica]: https://github.com/charmbracelet/harmonica
[termenv]: https://github.com/muesli/termenv
[reflow]: https://github.com/muesli/reflow

## Bubble Tea in the Wild

For some Bubble Tea programs in production, see:

* [AT CLI](https://github.com/daskycodes/at_cli): a utility for executing AT Commands via serial port connections
* [Canard](https://github.com/mrusme/canard): an RSS client
* [The Charm Tool](https://github.com/charmbracelet/charm): the Charm user account manager
* [clidle](https://github.com/ajeetdsouza/clidle): a Wordle clone for your terminal
* [fm](https://github.com/knipferrc/fm): a terminal-based file manager
* [fork-cleaner](https://github.com/caarlos0/fork-cleaner): cleans up old and inactive forks in your GitHub account
* [gambit](https://github.com/maaslalani/gambit): play chess in the terminal
* [gembro](https://git.sr.ht/~rafael/gembro): a mouse-driven Gemini browser
* [gh-b](https://github.com/joaom00/gh-b): GitHub CLI extension to easily manage your branches
* [gh-dash](https://www.github.com/dlvhdr/gh-dash): GitHub cli extension to display a dashboard of PRs and issues
* [gitflow-toolkit](https://github.com/mritd/gitflow-toolkit): a GitFlow submission tool
* [Glow](https://github.com/charmbracelet/glow): a markdown reader, browser and online markdown stash
* [gocovsh](https://github.com/orlangure/gocovsh): explore Go coverage reports from the CLI
* [httpit](https://github.com/gonetx/httpit): a rapid http(s) benchmark tool
* [IDNT](https://github.com/r-darwish/idnt): batch software uninstaller
* [kboard](https://github.com/CamiloGarciaLaRotta/kboard): a typing game
* [mergestat](https://github.com/mergestat/mergestat): run SQL queries on git repositories
* [mc](https://github.com/minio/mc): the official [MinIO](https://min.io) client
* [portal][portal]: securely send transfer between computers
* [redis-viewer](https://github.com/SaltFishPr/redis-viewer): browse Redis databases
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
* [Typer](https://github.com/maaslalani/typer): a typing test
* [wishlist](https://github.com/charmbracelet/wishlist): an SSH directory

[portal]: https://github.com/ZinoKader/portal

## Feedback

We'd love to hear your thoughts on this tutorial. Feel free to drop us a note!

* [Twitter](https://twitter.com/charmcli)
* [The Fediverse](https://mastodon.technology/@charm)

## Acknowledgments

Bubble Tea is based on the paradigms of [The Elm Architecture][elm] by Evan
Czaplicki et alia and the excellent [go-tea][gotea] by TJ Holowaychuk.

[elm]: https://guide.elm-lang.org/architecture/
[gotea]: https://github.com/tj/go-tea

## License

[MIT](https://github.com/charmbracelet/bubbletea/raw/master/LICENSE)

***

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400"></a>

Charm热爱开源 • Charm loves open source
