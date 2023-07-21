package tea

import (
	"context"
	"io"

	"github.com/muesli/termenv"
)

// WithContext lets you specify a context in which to run the Program. This is
// useful if you want to cancel the execution from outside. When a Program gets
// cancelled it will exit with an error ErrProgramKilled.
func (p *Program[M]) WithContext(ctx context.Context) *Program[M] {
	p.ctx = ctx
	return p
}

// WithOutput sets the output which, by default, is stdout. In most cases you
// won't need to use this.
func (p *Program[M]) WithOutput(output io.Writer) *Program[M] {
	if o, ok := output.(*termenv.Output); ok {
		p.output = o
	} else {
		p.output = termenv.NewOutput(output, termenv.WithColorCache(true))
	}
	return p
}

// WithInput sets the input which, by default, is stdin. In most cases you
// won't need to use this. To disable input entirely pass nil.
//
//	p := NewProgram(model).WithInput(nil)
func (p *Program[M]) WithInput(input io.Reader) *Program[M] {
	p.input = input
	p.inputType = customInput
	return p
}

// WithInputTTY opens a new TTY for input (or console input device on Windows).
func (p *Program[M]) WithInputTTY() *Program[M] {
	p.inputType = ttyInput
	return p
}

// WithoutSignalHandler disables the signal handler that Bubble Tea sets up for
// Programs. This is useful if you want to handle signals yourself.
func (p *Program[M]) WithoutSignalHandler() *Program[M] {
	p.startupOptions |= withoutSignalHandler
	return p
}

// WithoutCatchPanics disables the panic catching that Bubble Tea does by
// default. If panic catching is disabled the terminal will be in a fairly
// unusable state after a panic because Bubble Tea will not perform its usual
// cleanup on exit.
func (p *Program[M]) WithoutCatchPanics() *Program[M] {
	p.startupOptions |= withoutCatchPanics
	return p
}

// WithoutSignals will ignore OS signals.
// This is mainly useful for testing.
func (p *Program[M]) WithoutSignals() *Program[M] {
	p.ignoreSignals = true
	return p
}

// WithAltScreen starts the program with the alternate screen buffer enabled
// (i.e. the program starts in full window mode). Note that the altscreen will
// be automatically exited when the program quits.
//
// To enter the altscreen once the program has already started running use the
// EnterAltScreen command.
func (p *Program[M]) WithAltScreen() *Program[M] {
	p.startupOptions |= withAltScreen
	return p
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
func (p *Program[M]) WithMouseCellMotion() *Program[M] {
	p.startupOptions |= withMouseCellMotion // set
	p.startupOptions &^= withMouseAllMotion // clear
	return p
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
func (p *Program[M]) WithMouseAllMotion() *Program[M] {
	p.startupOptions |= withMouseAllMotion   // set
	p.startupOptions &^= withMouseCellMotion // clear
	return p
}

// WithoutRenderer disables the renderer. When this is set output and log
// statements will be plainly sent to stdout (or another output if one is set)
// without any rendering and redrawing logic. In other words, printing and
// logging will behave the same way it would in a non-TUI commandline tool.
// This can be useful if you want to use the Bubble Tea framework for a non-TUI
// application, or to provide an additional non-TUI mode to your Bubble Tea
// programs. For example, your program could behave like a daemon if output is
// not a TTY.
func (p *Program[M]) WithoutRenderer() *Program[M] {
	p.renderer = &nilRenderer{}
	return p
}

// WithANSICompressor removes redundant ANSI sequences to produce potentially
// smaller output, at the cost of some processing overhead.
//
// This feature is provisional, and may be changed or removed in a future version
// of this package.
func (p *Program[M]) WithANSICompressor() *Program[M] {
	p.startupOptions |= withANSICompressor
	return p
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
//	tea.NewProgram(Model{}).
//		.WithFilter(func (m tea.Model, msg tea.Msg) tea.Msg {
//			if _, ok := msg.(tea.QuitMsg); !ok {
//				return msg
//			}
//
//			model := m.(myModel)
//			if model.hasChanges {
//				return nil
//			}
//
//			return msg
//		});
func (p *Program[M]) WithFilter(filter func(M, Msg) Msg) *Program[M] {
	p.filter = filter
	return p
}

// WithMaxFPS sets a custom maximum FPS at which the renderer should run. If
// less than 1, the default value of 60 will be used. If over 120, the FPS
// will be capped at 120.
func (p *Program[M]) WithFPS(fps int) *Program[M] {
	p.fps = fps
	return p
}
