//go:build wasip1 || wasip2

package tea

// initInput on WASI: there is no raw mode and no /dev/tty; the program
// runs with non-TTY (pipe) input, which Bubble Tea supports natively.
func (p *Program) initInput() (err error) { return nil }

// suspendSupported: no SIGTSTP or process groups on WASI.
const suspendSupported = false

func suspendProcess() {}
