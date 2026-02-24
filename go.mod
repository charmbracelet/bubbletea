module charm.land/bubbletea/v2

retract v2.0.0-beta1 // We add a "." after the "beta" in the version number.

go 1.24.2

toolchain go1.24.4

require (
	github.com/charmbracelet/colorprofile v0.4.1
	github.com/charmbracelet/ultraviolet v0.0.0-20260205113103-524a6607adb8
	github.com/charmbracelet/x/ansi v0.11.6
	github.com/charmbracelet/x/exp/golden v0.0.0-20241212170349-ad4b7ae0f25f
	github.com/charmbracelet/x/term v0.2.2
	github.com/lucasb-eyer/go-colorful v1.3.0
	github.com/muesli/cancelreader v0.2.2
	golang.org/x/sys v0.40.0
)

require (
	github.com/aymanbagabas/go-udiff v0.2.0 // indirect
	github.com/charmbracelet/x/termios v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.9.0 // indirect
	github.com/clipperhouse/stringish v0.1.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.5.0 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/sync v0.19.0 // indirect
)
