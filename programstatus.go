package tea

import uv "github.com/charmbracelet/ultraviolet"

// ProgramStatus is a Program Status Protocol (OSC 7501) report. It tells the
// terminal whether the program is idle, working, waiting on the user, done,
// or failed, and why. Set it on [View.ProgramStatus].
//
// Support depends on the terminal. Terminals that don't support it ignore it.
//
// See: https://www.superlogical.com/rex/docs/build/program-status
type ProgramStatus = uv.ProgramStatus

// ProgramState is the state of a [ProgramStatus].
type ProgramState = uv.ProgramState

// ProgramStatusKind says what a blocked program waits for.
type ProgramStatusKind = uv.ProgramStatusKind

// Program states.
const (
	ProgramStateIdle    = uv.ProgramStateIdle
	ProgramStateWorking = uv.ProgramStateWorking
	ProgramStateDone    = uv.ProgramStateDone
	ProgramStateBlocked = uv.ProgramStateBlocked
	ProgramStateError   = uv.ProgramStateError
	ProgramStateClear   = uv.ProgramStateClear
)

// Program status kinds.
const (
	ProgramStatusKindPermission = uv.ProgramStatusKindPermission
	ProgramStatusKindQuestion   = uv.ProgramStatusKindQuestion
	ProgramStatusKindAuth       = uv.ProgramStatusKindAuth
)

// ProgramStatusSupportMsg is sent when the terminal replies to
// [RequestProgramStatusSupport], meaning it supports the Program Status
// Protocol.
type ProgramStatusSupportMsg struct{}

// programStatusSupportMsg is an internal message that queries the terminal
// for Program Status Protocol support.
type programStatusSupportMsg struct{}

// RequestProgramStatusSupport is a command that asks the terminal whether it
// supports the Program Status Protocol (OSC 7501). A supporting terminal
// replies with a [ProgramStatusSupportMsg]. Terminals that don't support it
// never reply, so pick your own timeout. Use [WithInput] to enable input,
// since the reply cannot be read otherwise.
//
// Detection is optional: reports can be sent without it.
func RequestProgramStatusSupport() Msg {
	return programStatusSupportMsg{}
}
