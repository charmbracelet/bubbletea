package tea

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/charmbracelet/colorprofile"
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
		p.externalCtx = ctx
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
		p.disableInput = input == nil
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
		p.disableSignalHandler = true
	}
}

// WithoutCatchPanics disables the panic catching that Bubble Tea does by
// default. If panic catching is disabled the terminal will be in a fairly
// unusable state after a panic because Bubble Tea will not perform its usual
// cleanup on exit.
func WithoutCatchPanics() ProgramOption {
	return func(p *Program) {
		p.disableCatchPanics = true
	}
}

// WithoutSignals will ignore OS signals.
// This is mainly useful for testing.
func WithoutSignals() ProgramOption {
	return func(p *Program) {
		atomic.StoreUint32(&p.ignoreSignals, 1)
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
		p.disableRenderer = true
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

// WithColorProfile sets the color profile that the program will use. This is
// useful when you want to force a specific color profile. By default, Bubble
// Tea will try to detect the terminal's color profile from environment
// variables and terminfo capabilities. Use [tea.WithEnvironment] to set custom
// environment variables.
func WithColorProfile(profile colorprofile.Profile) ProgramOption {
	return func(p *Program) {
		p.profile = &profile
	}
}

// WithWindowSize sets the initial size of the terminal window. This is useful
// when you need to set the initial size of the terminal window, for example
// during testing or when you want to run your program in a non-interactive
// environment.
func WithWindowSize(width, height int) ProgramOption {
	return func(p *Program) {
		p.width = width
		p.height = height
	}
}
