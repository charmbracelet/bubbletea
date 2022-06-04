package tea

// WindowSizeSource is used to specify which file descriptor to be used when
// determining the window size.
// Either output or input can be used.
type WindowSizeSource int

const (
	WindowSizeSourceOutput WindowSizeSource = iota
	WindowSizeSourceInput
)

func (w WindowSizeSource) String() string {
	switch w {
	case WindowSizeSourceOutput:
		return "output"
	case WindowSizeSourceInput:
		return "input"
	default:
		return ""
	}
}
