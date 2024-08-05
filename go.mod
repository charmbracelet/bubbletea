module github.com/charmbracelet/bubbletea

go 1.18

require (
	github.com/charmbracelet/lipgloss v0.12.1
	github.com/charmbracelet/shampoo v0.0.0-00010101000000-000000000000
	github.com/charmbracelet/x/ansi v0.1.5-0.20240805145438-99a7cad109d8
	github.com/charmbracelet/x/input v0.1.3
	github.com/charmbracelet/x/term v0.1.1
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6
	github.com/muesli/cancelreader v0.2.2
	golang.org/x/sync v0.7.0
	golang.org/x/sys v0.22.0
)

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/x/windows v0.1.2 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
)

replace github.com/rivo/uniseg => github.com/aymanbagabas/uniseg v0.4.8-0.20240530203522-35d7fd3fe5ce

replace github.com/charmbracelet/shampoo => ../shampoo
