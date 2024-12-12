package tea

import (
	"fmt"

	"github.com/charmbracelet/colorprofile"
)

const (
	// defaultFramerate specifies the maximum interval at which we should
	// update the view.
	defaultFPS = 60
	maxFPS     = 120
)

// renderer is the interface for Bubble Tea renderers.
type renderer interface {
	// close closes the renderer and flushes any remaining data.
	close() error

	// render renders a frame to the output.
	render(string)

	// flush flushes the renderer's buffer to the output.
	flush() error

	// reset resets the renderer's state to its initial state.
	reset()

	// insertAbove inserts unmanaged lines above the renderer.
	insertAbove(string)

	// enterAltScreen enters the alternate screen buffer.
	enterAltScreen()

	// exitAltScreen exits the alternate screen buffer.
	exitAltScreen()

	// showCursor shows the cursor.
	showCursor()

	// hideCursor hides the cursor.
	hideCursor()

	// resize notify the renderer of a terminal resize.
	resize(int, int)

	// setColorProfile sets the color profile.
	setColorProfile(colorprofile.Profile)

	// moveTo moves the cursor to the given position.
	moveTo(int, int)

	// clearScreen clears the screen.
	clearScreen()

	// repaint forces a full repaint.
	repaint()
}

// repaintMsg forces a full repaint.
type repaintMsg struct{}

type printLineMessage struct {
	messageBody string
}

// Println prints above the Program. This output is unmanaged by the program and
// will persist across renders by the Program.
//
// Unlike fmt.Println (but similar to log.Println) the message will be print on
// its own line.
//
// If the altscreen is active no output will be printed.
func Println(args ...interface{}) Cmd {
	return func() Msg {
		return printLineMessage{
			messageBody: fmt.Sprint(args...),
		}
	}
}

// Printf prints above the Program. It takes a format template followed by
// values similar to fmt.Printf. This output is unmanaged by the program and
// will persist across renders by the Program.
//
// Unlike fmt.Printf (but similar to log.Printf) the message will be print on
// its own line.
//
// If the altscreen is active no output will be printed.
func Printf(template string, args ...interface{}) Cmd {
	return func() Msg {
		return printLineMessage{
			messageBody: fmt.Sprintf(template, args...),
		}
	}
}
