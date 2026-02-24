package tea

import (
	"bytes"
	"image/color"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/exp/golden"
)

type testViewOpts struct {
	altScreen   bool
	mouseMode   MouseMode
	showCursor  bool
	disableBp   bool
	keyReleases bool
	bgColor     color.Color
}

func testViewOptsCmds(opts ...testViewOpts) []Cmd {
	cmds := make([]Cmd, len(opts))
	for i, o := range opts {
		o := o
		cmds[i] = func() Msg {
			return o
		}
	}
	return cmds
}

type testViewModel struct {
	*testModel
	opts testViewOpts
}

func (m *testViewModel) Update(msg Msg) (Model, Cmd) {
	switch msg := msg.(type) {
	case testViewOpts:
		m.opts = msg
		return m, nil
	}
	tm, cmd := m.testModel.Update(msg)
	m.testModel = tm.(*testModel)
	return m, cmd
}

func (m *testViewModel) View() View {
	v := m.testModel.View()
	v.AltScreen = m.opts.altScreen
	v.MouseMode = m.opts.mouseMode
	v.DisableBracketedPasteMode = m.opts.disableBp
	v.KeyboardEnhancements.ReportEventTypes = m.opts.keyReleases
	v.BackgroundColor = m.opts.bgColor
	if m.opts.showCursor {
		v.Cursor = NewCursor(0, 0)
	}
	return v
}

func TestViewModel(t *testing.T) {
	tests := []struct {
		name string
		opts []testViewOpts
	}{
		{
			name: "altscreen",
			opts: []testViewOpts{
				{altScreen: true},
				{altScreen: false},
			},
		},
		{
			name: "altscreen_autoexit",
			opts: []testViewOpts{
				{altScreen: true},
			},
		},
		{
			name: "mouse_cellmotion",
			opts: []testViewOpts{
				{mouseMode: MouseModeCellMotion},
			},
		},
		{
			name: "mouse_allmotion",
			opts: []testViewOpts{
				{mouseMode: MouseModeAllMotion},
			},
		},
		{
			name: "mouse_disable",
			opts: []testViewOpts{
				{mouseMode: MouseModeAllMotion},
				{mouseMode: MouseModeNone},
			},
		},
		{
			name: "cursor_hide",
			opts: []testViewOpts{
				{},
			},
		},
		{
			name: "cursor_hideshow",
			opts: []testViewOpts{
				{showCursor: false},
				{showCursor: true},
			},
		},
		{
			name: "bp_stop_start",
			opts: []testViewOpts{
				{disableBp: true},
				{disableBp: false},
			},
		},
		{
			name: "kitty_stop_startreleases",
			opts: []testViewOpts{
				{},
				{keyReleases: true},
			},
		},
		{
			name: "bg_set_color",
			opts: []testViewOpts{
				{bgColor: color.RGBA{255, 255, 255, 255}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			var in bytes.Buffer

			m := &testViewModel{testModel: &testModel{}}
			p := NewProgram(m,
				// Set the initial window size for the program.
				WithWindowSize(80, 24),
				// Use ANSI256 to increase test coverage.
				WithColorProfile(colorprofile.ANSI256),
				// always use xterm and 256 colors for tests
				WithEnvironment([]string{"TERM=xterm-256color"}),
				WithInput(&in),
				WithOutput(&buf),
			)

			go p.Send(append(sequenceMsg(testViewOptsCmds(test.opts...)), Quit))

			if _, err := p.Run(); err != nil {
				t.Fatal(err)
			}
			golden.RequireEqual(t, buf.Bytes())
		})
	}
}

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
			name: "read_set_clipboard",
			cmds: []Cmd{ReadClipboard, SetClipboard("success")},
		},
		{
			name: "bg_fg_cur_color",
			cmds: []Cmd{RequestForegroundColor, RequestBackgroundColor, RequestCursorColor},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			var in bytes.Buffer

			m := &testModel{}
			p := NewProgram(m,
				// Set the initial window size for the program.
				WithWindowSize(80, 24),
				// Use ANSI256 to increase test coverage.
				WithColorProfile(colorprofile.ANSI256),
				// always use xterm and 256 colors for tests
				WithEnvironment([]string{"TERM=xterm-256color"}),
				WithInput(&in),
				WithOutput(&buf),
			)

			go p.Send(append(test.cmds, Quit))

			if _, err := p.Run(); err != nil {
				t.Fatal(err)
			}
			golden.RequireEqual(t, buf.Bytes())
		})
	}
}
