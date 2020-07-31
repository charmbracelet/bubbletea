Bubble Tea
==========

<p>
    <img src="https://stuff.charm.sh/bubble-tea-header-github.png" width="523" height"225" alt="Bubble Tea Title Treatment"><br>
    <a href="https://pkg.go.dev/github.com/charmbracelet/bubbletea?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc"></a>
    <a href="https://github.com/charmbracelet/bubbletea/actions"><img src="https://github.com/charmbracelet/glow/workflows/build/badge.svg" alt="Build Status"></a>
</p>

The fun, functional and stateful way to build terminal apps. A Go framework
based on [The Elm Architecture][elm].

Bubble Tea is well-suited for simple and complex terminal applications,
either inline, full-window, or a mix of both. It's been battle-tested in
several large projects and is production-ready.

It features a standard framerate-based renderer which is used by default as
well as a renderer for high-performance scrollable regions, which works
alongside the main renderer.

To get started, see the [tutorials][tutorials] and [examples][examples].

[tutorials]: https://github.com/charmbracelet/tea/tree/master/tutorials
[examples]: https://github.com/charmbracelet/tea/tree/master/examples


## Bubble Tea in the Wild

For some Bubble Tea programs in production, see:

* [Glow](https://github.com/charmbracelet/glow): a markdown reader, browser and online markdown stash
* [The Charm Tool](https://github.com/charmbracelet/charm): the Charm user account manager


## Libraries we use with Bubble Tea

* [Bubbles][bubbles] various Bubble Tea components we've built
* [Termenv][termenv]: Advanced ANSI styling for terminal applications
* [Reflow][reflow]: ANSI-aware methods for reflowing blocks of text
* [go-runewidth][runewidth]: Get the physical width of strings in terms of terminal cells. Many runes, such as East Asian charcters and emojis, are two cells wide, so measuring a layout with `len()` often won't cut it!

[termenv]: https://github.com/muesli/termenv
[reflow]: https://github.com/muesli/reflow
[bubbles]: https://github.com/charmbracelet/bubbles
[runewidth]: https://github.com/mattn/go-runewidth


## Acknowledgments

Based on [The Elm Architecture][elm] by Evan Czaplicki et alia
and [go-tea][gotea] by TJ Holowaychuk.

[elm]: https://guide.elm-lang.org/architecture/
[gotea]: https://github.com/tj/go-tea


## License

[MIT](https://github.com/charmbracelet/bubbletea/raw/master/LICENSE)


***

A [Charm](https://charm.sh) project.

<img alt="the Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400">

Charm热爱开源!
