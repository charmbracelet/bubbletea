package tea

import (
	"bytes"
	"testing"
)

func TestClearMsg(t *testing.T) {
	tests := []struct {
		name     string
		cmds     sequenceMsg
		expected string
	}{
		{
			name:     "clear_screen",
			cmds:     []Cmd{ClearScreen},
			expected: "\x1b[?25l\x1b[?2004h\x1b[2J\x1b[1;1H\rsuccess\r\n\x1b[D\x1b[2K\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "altscreen",
			cmds:     []Cmd{EnterAltScreen, ExitAltScreen},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1049h\x1b[2J\x1b[1;1H\x1b[?25l\x1b[?1049l\x1b[?25l\rsuccess\r\n\x1b[D\x1b[2K\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "altscreen_autoexit",
			cmds:     []Cmd{EnterAltScreen},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1049h\x1b[2J\x1b[1;1H\x1b[?25l\rsuccess\r\n\x1b[2;0H\x1b[2K\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l\x1b[?1049l\x1b[?25h",
		},
		{
			name:     "mouse_cellmotion",
			cmds:     []Cmd{EnableMouseCellMotion},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1002h\x1b[?1006h\rsuccess\r\n\x1b[D\x1b[2K\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "mouse_allmotion",
			cmds:     []Cmd{EnableMouseAllMotion},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1003h\x1b[?1006h\rsuccess\r\n\x1b[D\x1b[2K\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "mouse_disable",
			cmds:     []Cmd{EnableMouseAllMotion, DisableMouse},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1003h\x1b[?1006h\x1b[?1002l\x1b[?1003l\x1b[?1006l\rsuccess\r\n\x1b[D\x1b[2K\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "cursor_hide",
			cmds:     []Cmd{HideCursor},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?25l\rsuccess\r\n\x1b[D\x1b[2K\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "cursor_hideshow",
			cmds:     []Cmd{HideCursor, ShowCursor},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?25l\x1b[?25h\rsuccess\r\n\x1b[D\x1b[2K\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "bp_stop_start",
			cmds:     []Cmd{DisableBracketedPaste, EnableBracketedPaste},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?2004l\x1b[?2004h\rsuccess\r\n\x1b[D\x1b[2K\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			var in bytes.Buffer

			m := &testModel{}
			p := NewProgram(m, WithInput(&in), WithOutput(&buf))

			test.cmds = append(test.cmds, Quit)
			go p.Send(test.cmds)

			if _, err := p.Run(); err != nil {
				t.Fatal(err)
			}

			if buf.String() != test.expected {
				t.Errorf("expected embedded sequence:\n%q\ngot:\n%q", test.expected, buf.String())
			}
		})
	}
}
