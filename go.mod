module github.com/charmbracelet/bubbletea/v2

retract v2.0.0-beta1 // We add a "." after the "beta" in the version number.

go 1.24.0

toolchain go1.24.4

require (
	github.com/charmbracelet/colorprofile v0.3.1
	github.com/charmbracelet/ultraviolet v0.0.0-20250721205647-f6ac6eda5d42
	github.com/charmbracelet/x/ansi v0.9.3
	github.com/charmbracelet/x/exp/golden v0.0.0-20241212170349-ad4b7ae0f25f
	github.com/charmbracelet/x/term v0.2.1
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/muesli/cancelreader v0.2.2
	golang.org/x/sync v0.16.0
	golang.org/x/sys v0.34.0
)

require (
	github.com/aymanbagabas/go-udiff v0.2.0 // indirect
	github.com/charmbracelet/x/termios v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.2.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
)
