package tea

import (
	"bytes"
	"image/color"
	"runtime"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/exp/golden"
)

func TestClearMsg(t *testing.T) {
	type test struct {
		name string
		cmds sequenceMsg
	}
	tests := []test{
		{
			name: "clear_screen",
			cmds: []Cmd{ClearScreen},
		},
		{
			name: "altscreen",
			cmds: []Cmd{EnterAltScreen, ExitAltScreen},
		},
		{
			name: "altscreen_autoexit",
			cmds: []Cmd{EnterAltScreen},
		},
		{
			name: "mouse_cellmotion",
			cmds: []Cmd{EnableMouseCellMotion},
		},
		{
			name: "mouse_allmotion",
			cmds: []Cmd{EnableMouseAllMotion},
		},
		{
			name: "mouse_disable",
			cmds: []Cmd{EnableMouseAllMotion, DisableMouse},
		},
		{
			name: "cursor_hide",
			cmds: []Cmd{HideCursor},
		},
		{
			name: "cursor_hideshow",
			cmds: []Cmd{HideCursor, ShowCursor},
		},
		{
			name: "bp_stop_start",
			cmds: []Cmd{DisableBracketedPaste, EnableBracketedPaste},
		},
		{
			name: "read_set_clipboard",
			cmds: []Cmd{ReadClipboard, SetClipboard("success")},
		},
		{
			name: "bg_fg_cur_color",
			cmds: []Cmd{RequestForegroundColor, RequestBackgroundColor, RequestCursorColor},
		},
		{
			name: "bg_set_color",
			cmds: []Cmd{SetBackgroundColor(color.RGBA{255, 255, 255, 255})},
		},
		{
			name: "grapheme_clustering",
			cmds: []Cmd{EnableGraphemeClustering},
		},
	}

	if runtime.GOOS == "windows" {
		// Windows supports enhanced keyboard features through the Windows API, not through ANSI sequences.
		tests = append(tests, test{
			name: "kitty_start_windows",
			cmds: []Cmd{DisableKeyboardEnhancements, EnableKeyboardEnhancements(WithKeyReleases)},
		})
	} else {
		tests = append(tests, test{
			name: "kitty_start_other",
			cmds: []Cmd{DisableKeyboardEnhancements, EnableKeyboardEnhancements(WithKeyReleases)},
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			var in bytes.Buffer

			m := &testModel{}
			p := NewProgram(m, WithInput(&in), WithOutput(&buf),
				WithEnvironment([]string{
					"TERM=xterm-256color", // always use xterm and 256 colors for tests
				}),
				// Use ANSI256 to increase test coverage.
				WithColorProfile(colorprofile.ANSI256))

			// Set the initial window size for the program.
			p.width, p.height = 80, 24

			go p.Send(append(test.cmds, Quit))

			if _, err := p.Run(); err != nil {
				t.Fatal(err)
			}
			golden.RequireEqual(t, buf.Bytes())
		})
	}
}
