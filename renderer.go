package tea

import (
	"fmt"
	"image/color"

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
	render(View)

	// hit returns possible hit messages for the renderer.
	hit(MouseMsg) []Msg

	// flush flushes the renderer's buffer to the output.
	flush(*Program) error

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

	// setCursorColor sets the terminal's cursor color.
	setCursorColor(color.Color)

	// setForegroundColor sets the terminal's foreground color.
	setForegroundColor(color.Color)

	// setBackgroundColor sets the terminal's background color.
	setBackgroundColor(color.Color)

	// setWindowTitle sets the terminal window title.
	setWindowTitle(string)

	// resize notify the renderer of a terminal resize.
	resize(int, int)

	// setColorProfile sets the color profile.
	setColorProfile(colorprofile.Profile)

	// clearScreen clears the screen.
	clearScreen()

	// repaint forces a full repaint.
	repaint()

	writeString(string) (int, error)
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
