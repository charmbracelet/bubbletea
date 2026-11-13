//go:build windows
// +build windows

package tea

import "golang.org/x/sys/windows"

// drainInput discards any pending console input events to remove unsolicited
// terminal responses that arrived after the input reader was cancelled.
// Without this, those bytes can be read by the user's shell after exit and
// printed as garbage characters.
func (p *Program) drainInput() {
	if p.ttyInput == nil {
		return
	}
	_ = windows.FlushConsoleInputBuffer(windows.Handle(p.ttyInput.Fd()))
}
