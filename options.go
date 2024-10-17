package tea

import (
	"context"
	"io"
	"sync/atomic"
)

// ProgramOption is used to set options when initializing a Program. Program can
// accept a variable number of options.
//
// Example usage:
//
//	p := NewProgram(model, WithInput(someInput), WithOutput(someOutput))
type ProgramOption func(*Program)

// WithContext lets you specify a context in which to run the Program. This is
// useful if you want to cancel the execution from outside. When a Program gets
// cancelled it will exit with an error ErrProgramKilled.
func WithContext(ctx context.Context) ProgramOption {
	return func(p *Program) {
		p.ctx = ctx
	}
}

// WithOutput sets the output which, by default, is stdout. In most cases you
// won't need to use this.
func WithOutput(output io.Writer) ProgramOption {
	return func(p *Program) {
		p.output = output
	}
}

// WithInput sets the input which, by default, is stdin. In most cases you
// won't need to use this. To disable input entirely pass nil.
//
//	p := NewProgram(model, WithInput(nil))
func WithInput(input io.Reader) ProgramOption {
	return func(p *Program) {
		p.input = input
		p.inputType = customInput
	}
}

// WithInputTTY opens a new TTY for input (or console input device on Windows).
func WithInputTTY() ProgramOption {
	return func(p *Program) {
		p.inputType = ttyInput
	}
}

// WithEnvironment sets the environment variables that the program will use.
// This useful when the program is running in a remote session (e.g. SSH) and
// you want to pass the environment variables from the remote session to the
// program.
//
// Example:
//
//	var sess ssh.Session // ssh.Session is a type from the github.com/charmbracelet/ssh package
//	pty, _, _ := sess.Pty()
//	environ := append(sess.Environ(), "TERM="+pty.Term)
//	p := tea.NewProgram(model, tea.WithEnvironment(environ)
func WithEnvironment(env []string) ProgramOption {
	return func(p *Program) {
		p.environ = env
	}
}

// WithoutSignalHandler disables the signal handler that Bubble Tea sets up for
// Programs. This is useful if you want to handle signals yourself.
func WithoutSignalHandler() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withoutSignalHandler
	}
}

// WithoutCatchPanics disables the panic catching that Bubble Tea does by
// default. If panic catching is disabled the terminal will be in a fairly
// unusable state after a panic because Bubble Tea will not perform its usual
// cleanup on exit.
func WithoutCatchPanics() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withoutCatchPanics
	}
}

// WithoutSignals will ignore OS signals.
// This is mainly useful for testing.
func WithoutSignals() ProgramOption {
	return func(p *Program) {
		atomic.StoreUint32(&p.ignoreSignals, 1)
	}
}

// WithAltScreen starts the program with the alternate screen buffer enabled
// (i.e. the program starts in full window mode). Note that the altscreen will
// be automatically exited when the program quits.
//
// Example:
//
//	p := tea.NewProgram(Model{}, tea.WithAltScreen())
//	if _, err := p.Run(); err != nil {
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

// WithoutBracketedPaste starts the program with bracketed paste disabled.
func WithoutBracketedPaste() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withoutBracketedPaste
	}
}

// WithMouseCellMotion starts the program with the mouse enabled in "cell
// motion" mode.
//
// Cell motion mode enables mouse click, release, and wheel events. Mouse
// movement events are also captured if a mouse button is pressed (i.e., drag
// events). Cell motion mode is better supported than all motion mode.
//
// This will try to enable the mouse in extended mode (SGR), if that is not
// supported by the terminal it will fall back to normal mode (X10).
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
// This will try to enable the mouse in extended mode (SGR), if that is not
// supported by the terminal it will fall back to normal mode (X10).
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
	return func(p *Program) {
		p.renderer = &nilRenderer{}
	}
}

// WithANSICompressor removes redundant ANSI sequences to produce potentially
// smaller output, at the cost of some processing overhead.
//
// This feature is provisional, and may be changed or removed in a future version
// of this package.
//
// Deprecated: this incurs a noticable performance hit. A future release will
// optimize ANSI automatically without the performance penalty.
func WithANSICompressor() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withANSICompressor
	}
}

// WithFilter supplies an event filter that will be invoked before Bubble Tea
// processes a tea.Msg. The event filter can return any tea.Msg which will then
// get handled by Bubble Tea instead of the original event. If the event filter
// returns nil, the event will be ignored and Bubble Tea will not process it.
//
// As an example, this could be used to prevent a program from shutting down if
// there are unsaved changes.
//
// Example:
//
//	func filter(m tea.Model, msg tea.Msg) tea.Msg {
//		if _, ok := msg.(tea.QuitMsg); !ok {
//			return msg
//		}
//
//		model := m.(myModel)
//		if model.hasChanges {
//			return nil
//		}
//
//		return msg
//	}
//
//	p := tea.NewProgram(Model{}, tea.WithFilter(filter));
//
//	if _,err := p.Run(); err != nil {
//		fmt.Println("Error running program:", err)
//		os.Exit(1)
//	}
func WithFilter(filter func(Model, Msg) Msg) ProgramOption {
	return func(p *Program) {
		p.filter = filter
	}
}

// WithFPS sets a custom maximum FPS at which the renderer should run. If
// less than 1, the default value of 60 will be used. If over 120, the FPS
// will be capped at 120.
func WithFPS(fps int) ProgramOption {
	return func(p *Program) {
		p.fps = fps
	}
}

// WithReportFocus enables reporting when the terminal gains and loses
// focus. When this is enabled [FocusMsg] and [BlurMsg] messages will be sent
// to your Update method.
//
// Note that while most terminals and multiplexers support focus reporting,
// some do not. Also note that tmux needs to be configured to report focus
// events.
func WithReportFocus() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withReportFocus
	}
}
