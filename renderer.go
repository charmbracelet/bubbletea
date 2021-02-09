package tea

type renderer interface {
	start()
	stop()
	write(string)
	altScreen() bool
	setAltScreen(bool)
}
