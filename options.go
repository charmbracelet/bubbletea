package tea

import "io"

// ProgramOption is used to set options when initializing a Program. Program can
// accept a variable number of options.
//
// Example usage:
//
//	p := NewProgram(model, WithInput(someInput), WithOutput(someOutput))
type ProgramOption func(*Program)

// WithOutput sets the output which, by default, is stdout. In most cases you
// won't need to use this.
func WithOutput(output io.Writer) ProgramOption {
	return func(m *Program) {
		m.output = output
	}
}

// WithInput sets the input which, by default, is stdin. In most cases you
// won't need to use this.
func WithInput(input io.Reader) ProgramOption {
	return func(m *Program) {
		m.input = input
		m.startupOptions |= withCustomInput
	}
}

// WithInputTTY open a new TTY for input (or console input device on Windows).
func WithInputTTY() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withInputTTY
	}
}

// WithoutCatchPanics disables the panic catching that Bubble Tea does by
// default. If panic catching is disabled the terminal will be in a fairly
// unusable state after a panic because Bubble Tea will not perform its usual
// cleanup on exit.
func WithoutCatchPanics() ProgramOption {
	return func(m *Program) {
		m.CatchPanics = false
	}
}

// WithAltScreen starts the program with the alternate screen buffer enabled
// (i.e. the program starts in full window mode). Note that the altscreen will
// be automatically exited when the program quits.
//
// Example:
//
//	p := tea.NewProgram(Model{}, tea.WithAltScreen())
//	if err := p.Start(); err != nil {
//	    fmt.Println("Error running program:", err)
//	    os.Exit(1)
//	}
//
// To enter the altscreen once the program has already started running use the
// EnterAltScreen command.
func WithAltScreen() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withAltScreen
	}
}

// WithMouseCellMotion starts the program with the mouse enabled in "cell
// motion" mode.
//
// Cell motion mode enables mouse click, release, and wheel events. Mouse
// movement events are also captured if a mouse button is pressed (i.e., drag
// events). Cell motion mode is better supported than all motion mode.
//
// To enable mouse cell motion once the program has already started running use
// the EnableMouseCellMotion command. To disable the mouse when the program is
// running use the DisableMouse command.
//
// The mouse will be automatically disabled when the program exits.
func WithMouseCellMotion() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withMouseCellMotion // set
		p.startupOptions &^= withMouseAllMotion // clear
	}
}

// WithMouseAllMotion starts the program with the mouse enabled in "all motion"
// mode.
//
// EnableMouseAllMotion is a special command that enables mouse click, release,
// wheel, and motion events, which are delivered regardless of whether a mouse
// button is pressed, effectively enabling support for hover interactions.
//
// Many modern terminals support this, but not all. If in doubt, use
// EnableMouseCellMotion instead.
//
// To enable the mouse once the program has already started running use the
// EnableMouseAllMotion command. To disable the mouse when the program is
// running use the DisableMouse command.
//
// The mouse will be automatically disabled when the program exits.
func WithMouseAllMotion() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withMouseAllMotion   // set
		p.startupOptions &^= withMouseCellMotion // clear
	}
}

// WithoutRenderer disables the renderer. When this is set output and log
// statements will be plainly sent to stdout (or another output if one is set)
// without any rendering and redrawing logic. In other words, printing and
// logging will behave the same way it would in a non-TUI commandline tool.
// This can be useful if you want to use the Bubble Tea framework for a non-TUI
// application, or to provide an additional non-TUI mode to your Bubble Tea
// programs. For example, your program could behave like a daemon if output is
// not a TTY.
func WithoutRenderer() ProgramOption {
	return func(m *Program) {
		m.renderer = &nilRenderer{}
	}
}

// WithANSICompressor removes redundant ANSI sequences to produce potentially
// smaller output, at the cost of some processing overhead.
//
// This feature is provisional, and may be changed removed in a future version
// of this package.
func WithANSICompressor() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withANSICompressor
	}
}
