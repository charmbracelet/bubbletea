package tea

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClearMsg(t *testing.T) {
	for name, test := range map[string]struct {
		cmds     sequenceMsg
		expected string
	}{
		"clear_screen": {
			cmds:     []Cmd{ClearScreen},
			expected: "\x1b[?25l\x1b[2J\x1b[1;1H\x1b[1;1Hsuccess\r\n\x1b[0D\x1b[2K\x1b[?25h\x1b[?1002l\x1b[?1003l",
		},
		"altscreen": {
			cmds:     []Cmd{EnterAltScreen, ExitAltScreen},
			expected: "\x1b[?25l\x1b[?1049h\x1b[2J\x1b[1;1H\x1b[1;1H\x1b[?25l\x1b[?1049l\x1b[?25lsuccess\r\n\x1b[0D\x1b[2K\x1b[?25h\x1b[?1002l\x1b[?1003l",
		},
		"altscreen_autoexit": {
			cmds:     []Cmd{EnterAltScreen},
			expected: "\x1b[?25l\x1b[?1049h\x1b[2J\x1b[1;1H\x1b[1;1H\x1b[?25lsuccess\r\n\x1b[2;0H\x1b[2K\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1049l\x1b[?25h",
		},
		"mouse_cellmotion": {
			cmds:     []Cmd{EnableMouseCellMotion},
			expected: "\x1b[?25l\x1b[?1002hsuccess\r\n\x1b[0D\x1b[2K\x1b[?25h\x1b[?1002l\x1b[?1003l",
		},
		"mouse_allmotion": {
			cmds:     []Cmd{EnableMouseAllMotion},
			expected: "\x1b[?25l\x1b[?1003hsuccess\r\n\x1b[0D\x1b[2K\x1b[?25h\x1b[?1002l\x1b[?1003l",
		},
		"mouse_disable": {
			cmds:     []Cmd{EnableMouseAllMotion, DisableMouse},
			expected: "\x1b[?25l\x1b[?1003h\x1b[?1002l\x1b[?1003lsuccess\r\n\x1b[0D\x1b[2K\x1b[?25h\x1b[?1002l\x1b[?1003l",
		},
		"cursor_hide": {
			cmds:     []Cmd{HideCursor},
			expected: "\x1b[?25l\x1b[?25lsuccess\r\n\x1b[0D\x1b[2K\x1b[?25h\x1b[?1002l\x1b[?1003l",
		},
		"cursor_hideshow": {
			cmds:     []Cmd{HideCursor, ShowCursor},
			expected: "\x1b[?25l\x1b[?25l\x1b[?25hsuccess\r\n\x1b[0D\x1b[2K\x1b[?25h\x1b[?1002l\x1b[?1003l",
		},
	} {
		t.Run(name, func(t *testing.T) {
			var (
				in  bytes.Buffer
				out bytes.Buffer
			)
			p := NewProgram(&testModel{}).WithInput(&in).WithOutput(&out)

			go p.Send(append(test.cmds, Quit))

			_, err := p.Run()
			assert.NoError(t, err)

			assert.Equal(t, test.expected, out.String())
		})
	}
}
