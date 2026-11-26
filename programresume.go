package tea

import "github.com/charmbracelet/x/ansi"

// ProgramResume tells the terminal how to resume the program after the
// terminal restarts, using the Program Resume Protocol (OSC 7577). Set it on
// [View.ProgramResume].
//
// Args is the command that brings the program back, program first: a bare
// command name looked up on PATH, or an absolute path such as the one from
// [os.Executable]. Dir optionally sets the absolute directory to run it in.
//
// The terminal may run the command in a fresh environment, more than once, or
// not at all, so it should carry everything the program needs to pick up where
// it left off, such as a session ID or the path of a state file.
//
// See: https://www.superlogical.com/rex/docs/build/program-resume
type ProgramResume = ansi.ProgramResume
