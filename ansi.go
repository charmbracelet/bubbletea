package tea

import "fmt"

// Escape sequence
const esc = "\033["

// Fullscreen switches to the altscreen and clears the terminal. The former
// view can be restored with ExitFullscreen().
func Fullscreen() {
	fmt.Print(esc + "?1049h" + esc + "H")
}

// ExitFullscreen exits the altscreen and returns the former terminal view
func ExitFullscreen() {
	fmt.Print(esc + "?1049l")
}

// ClearScreen clears the visible portion of the terminal. Effectively, it
// fills the terminal with blank spaces.
func ClearScreen() {
	fmt.Printf(esc + "2J" + esc + "3J" + esc + "1;1H")
}

// Invert inverts the foreground and background colors of a given string
func Invert(s string) string {
	return esc + "7m" + s + esc + "0m"
}

// Hide the cursor
func hideCursor() {
	fmt.Printf(esc + "?25l")
}

// Show the cursor
func showCursor() {
	fmt.Printf(esc + "?25h")
}

// Move the cursor down a given number of lines and place it at the beginning
// of the line
func cursorNextLine(n int) {
	fmt.Printf(esc+"%dE", n)
}

// Move the cursor up a given number of lines and place it at the beginning of
// the line
func cursorPrevLine(n int) {
	fmt.Printf(esc+"%dF", n)
}

// Clear the current line
func clearLine() {
	fmt.Printf(esc + "2K")
}

// Clear a given number of lines
func clearLines(n int) {
	clearLine()
	for i := 0; i < n; i++ {
		cursorPrevLine(1)
		clearLine()
	}
}
