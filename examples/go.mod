module examples

go 1.13

replace github.com/charmbracelet/bubbletea => ../

require (
	github.com/charmbracelet/bubbles v0.5.0
	github.com/charmbracelet/bubbletea v0.10.3
	github.com/fogleman/ease v0.0.0-20170301025033-8da417bf1776
	github.com/mattn/go-runewidth v0.0.9
	github.com/muesli/termenv v0.7.0
)

replace github.com/charmbracelet/bubbles => ../../bubbles
