package tea

import (
	"fmt"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

const (
	// defaultFramerate specifies the maximum interval at which we should
	// update the view.
	defaultFPS = 60
	maxFPS     = 120
)

// renderer is the interface for Bubble Tea renderers.
type renderer interface {
	// start starts the renderer.
	start()

	// close closes the renderer and flushes any remaining data.
	close() error

	// render renders a frame to the output.
	render(View)

	// flush flushes the renderer's buffer to the output.
	flush(closing bool) error

	// reset resets the renderer's state to its initial state.
	reset()

	// insertAbove inserts unmanaged lines above the renderer.
	insertAbove(string) error

	// setSyncdUpdates sets whether to use synchronized updates.
	setSyncdUpdates(bool)

	// setWidthMethod sets the method for calculating the width of the terminal.
	setWidthMethod(ansi.Method)

	// resize notify the renderer of a terminal resize.
	resize(int, int)

	// setColorProfile sets the color profile.
	setColorProfile(colorprofile.Profile)

	// clearScreen clears the screen.
	clearScreen()

	// writeString writes a string to the renderer's output.
	writeString(string) (int, error)

	// onMouse handles a mouse event.
	onMouse(MouseMsg) Cmd
}

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
func Println(args ...any) Cmd {
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
func Printf(template string, args ...any) Cmd {
	return func() Msg {
		return printLineMessage{
			messageBody: fmt.Sprintf(template, args...),
		}
	}
}

// encodeCursorStyle returns the integer value for the given cursor style and
// blink state.
func encodeCursorStyle(style CursorShape, blink bool) int {
	// We're using the ANSI escape sequence values for cursor styles.
	// We need to map both [style] and [steady] to the correct value.
	style = (style * 2) + 1 //nolint:mnd
	if !blink {
		style++
	}
	return int(style)
}
