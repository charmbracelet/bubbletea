package tea

// renderer is the interface for Bubble Tea renderers.
type renderer interface {
	start()
	stop()
	write(string)
	altScreen() bool
	setAltScreen(bool)
}
