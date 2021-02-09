package tea

type nilRenderer struct{}

func (n nilRenderer) start()              {}
func (n nilRenderer) stop()               {}
func (n nilRenderer) write(v string)      {}
func (n nilRenderer) altScreen() bool     { return false }
func (n nilRenderer) setAltScreen(v bool) {}
