//go:build linux

package tea

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

type cursorTTYFailWriter struct {
	tty      *os.File
	original *term.State
	sawRaw   bool
}

func (w *cursorTTYFailWriter) Write([]byte) (int, error) {
	state, err := term.GetState(w.tty.Fd())
	if err == nil && !reflect.DeepEqual(state, w.original) {
		w.sawRaw = true
	}
	return 0, errors.New("terminal output failed")
}

func TestCursorQueryErrorRestoresTTY(t *testing.T) {
	// A dedicated PTY isolates termios changes from the developer's terminal.
	fd, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("PTY unavailable: %v", err)
	}
	master := os.NewFile(uintptr(fd), "cursor-test-master")
	defer master.Close()
	if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		t.Fatal(err)
	}
	index, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		t.Fatal(err)
	}
	slave, err := os.OpenFile("/dev/pts/"+strconv.Itoa(index), os.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer slave.Close()
	original, err := term.GetState(slave.Fd())
	if err != nil {
		t.Fatal(err)
	}
	out := &cursorTTYFailWriter{tty: slave, original: original}
	p := NewProgram(&cursorRegressionModel{}, WithInput(slave), WithOutput(out), WithoutSignals(), WithWindowSize(80, 24))
	if _, err := p.Run(); err == nil {
		t.Fatal("expected terminal write failure")
	}
	if !out.sawRaw {
		t.Fatal("fixture never observed raw mode; restoration assertion would be vacuous")
	}
	after, err := term.GetState(slave.Fd())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, original) {
		t.Fatal("TTY settings were not restored after query failure")
	}
}
