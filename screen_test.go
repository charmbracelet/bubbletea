package tea

import (
	"bytes"
	"image/color"
	"runtime"
	"testing"

	"github.com/charmbracelet/colorprofile"
)

func TestClearMsg(t *testing.T) {
	type test struct {
		name     string
		cmds     sequenceMsg
		expected string
	}
	tests := []test{
		{
			name:     "clear_screen",
			cmds:     []Cmd{ClearScreen},
			expected: "\x1b[?25l\x1b[?2004h\x1b[2J\x1b[H\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h",
		},
		{
			name:     "altscreen",
			cmds:     []Cmd{EnterAltScreen, ExitAltScreen},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1049h\x1b[2J\x1b[H\x1b[?25l\x1b[?1049l\x1b[?25l\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h",
		},
		{
			name:     "altscreen_autoexit",
			cmds:     []Cmd{EnterAltScreen},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1049h\x1b[2J\x1b[H\x1b[?25l\x1b[H\rsuccess\r\n\x1b[2;H\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?1049l\x1b[?25h",
		},
		{
			name:     "mouse_cellmotion",
			cmds:     []Cmd{EnableMouseCellMotion},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1002h\x1b[?1006h\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "mouse_allmotion",
			cmds:     []Cmd{EnableMouseAllMotion},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1003h\x1b[?1006h\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006l",
		},
		{
			name:     "mouse_disable",
			cmds:     []Cmd{EnableMouseAllMotion, DisableMouse},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?1003h\x1b[?1006h\x1b[?1002l\x1b[?1003l\x1b[?1006l\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h",
		},
		{
			name:     "cursor_hide",
			cmds:     []Cmd{HideCursor},
			expected: "\x1b[?25l\x1b[?2004h\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h",
		},
		{
			name:     "cursor_hideshow",
			cmds:     []Cmd{HideCursor, ShowCursor},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?25h\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l",
		},
		{
			name:     "bp_stop_start",
			cmds:     []Cmd{DisableBracketedPaste, EnableBracketedPaste},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?2004l\x1b[?2004h\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h",
		},
		{
			name:     "read_set_clipboard",
			cmds:     []Cmd{ReadClipboard, SetClipboard("success")},
			expected: "\x1b[?25l\x1b[?2004h\x1b]52;c;?\a\x1b]52;c;c3VjY2Vzcw==\a\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h",
		},
		{
			name:     "bg_fg_cur_color",
			cmds:     []Cmd{RequestForegroundColor, RequestBackgroundColor, RequestCursorColor},
			expected: "\x1b[?25l\x1b[?2004h\x1b]10;?\a\x1b]11;?\a\x1b]12;?\a\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h",
		},
		{
			name:     "bg_set_color",
			cmds:     []Cmd{SetBackgroundColor(color.RGBA{255, 255, 255, 255})},
			expected: "\x1b[?25l\x1b[?2004h\x1b]11;#ffffff\a\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b]111\a",
		},
		{
			name:     "grapheme_clustering",
			cmds:     []Cmd{EnableGraphemeClustering},
			expected: "\x1b[?25l\x1b[?2004h\x1b[?2027h\x1b[?2027$p\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b[?2027l",
		},
	}

	if runtime.GOOS == "windows" {
		// Windows supports enhanced keyboard features through the Windows API, not through ANSI sequences.
		tests = append(tests, test{
			name:     "kitty_start",
			cmds:     []Cmd{DisableKeyboardEnhancements, RequestKeyboardEnhancements(WithKeyReleases)},
			expected: "\x1b[?25l\x1b[?2004h\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h",
		})
	} else {
		tests = append(tests, test{
			name:     "kitty_start",
			cmds:     []Cmd{DisableKeyboardEnhancements, RequestKeyboardEnhancements(WithKeyReleases)},
			expected: "\x1b[?25l\x1b[?2004h\x1b[>4;1m\x1b[?4m\x1b[>3u\x1b[?u\rsuccess\r\n\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h",
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			var in bytes.Buffer

			m := &testModel{}
			p := NewProgram(m, WithInput(&in), WithOutput(&buf),
				// Use ANSI256 to increase test coverage.
				WithColorProfile(colorprofile.ANSI256))

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
