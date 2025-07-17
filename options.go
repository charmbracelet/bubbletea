package tea

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
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

// WithKeyReleases enables support for reporting key release events. This is
// useful for terminals that support the Kitty keyboard protocol "Report event
// types" progressive enhancement feature.
//
// Note that not all terminals support this feature. If the terminal does not
// support this feature, the program will not receive key release events.
func WithKeyReleases() ProgramOption {
	return func(p *Program) {
		p.requestedEnhancements.kittyFlags |= ansi.KittyReportEventTypes
		p.requestedEnhancements.keyReleases = true
	}
}

// WithUniformKeyLayout enables support for reporting key events as though they
// were on a PC-101 layout. This is useful for uniform key event reporting
// across different keyboard layouts. This is equivalent to the Kitty keyboard
// protocol "Report alternate keys" and "Report all keys as escape codes"
// progressive enhancement features.
//
// Note that not all terminals support this feature. If the terminal does not
// support this feature, the program will not receive key events in
// uniform layout format.
func WithUniformKeyLayout() ProgramOption {
	return func(p *Program) {
		p.requestedEnhancements.kittyFlags |= ansi.KittyReportAlternateKeys | ansi.KittyReportAllKeysAsEscapeCodes
	}
}

// WithGraphemeClustering disables grapheme clustering. This is useful if you
// want to disable grapheme clustering for your program.
//
// Grapheme clustering is a character width calculation method that accurately
// calculates the width of wide characters in a terminal. This is useful for
// properly rendering double width characters such as emojis and CJK
// characters.
//
// See https://mitchellh.com/writing/grapheme-clusters-in-terminals
func WithGraphemeClustering() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withGraphemeClustering
	}
}

// WithColorProfile sets the color profile that the program will use. This is
// useful when you want to force a specific color profile. By default, Bubble
// Tea will try to detect the terminal's color profile from environment
// variables and terminfo capabilities. Use [tea.WithEnvironment] to set custom
// environment variables.
func WithColorProfile(profile colorprofile.Profile) ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withColorProfile
		p.profile = profile
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

// WithoutKeyEnhancements disables all key enhancements. This is useful if you
// want to disable all key enhancements for your program and keep your program
// legacy compatible with older terminals.
func WithoutKeyEnhancements() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withoutKeyEnhancements
	}
}
