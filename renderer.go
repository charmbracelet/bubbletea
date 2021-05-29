package tea

// renderer is the interface for Bubble Tea renderers.
type renderer interface {
	start()
	stop()
	write(string)
	repaint()
	altScreen() bool
	setAltScreen(bool)
}
