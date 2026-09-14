package tea

import (
	"bytes"
	"image/color"
	"testing"
	"time"

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
			t.Parallel()
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
			t.Parallel()
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

func TestPrintAbove(t *testing.T) {
	tests := []struct {
		name      string
		altScreen bool
		cmds      sequenceMsg
	}{
		{
			name: "println_above_inline",
			cmds: sequenceMsg{PrintlnAbove("hello from above")},
		},
		{
			name: "printf_above_inline",
			cmds: sequenceMsg{PrintfAbove("formatted %s", "message")},
		},
		{
			name:      "println_above_altscreen",
			altScreen: true,
			cmds:      sequenceMsg{PrintlnAbove("persisted line")},
		},
		{
			name:      "println_dropped_altscreen",
			altScreen: true,
			cmds:      sequenceMsg{Println("this should not appear")},
		},
		{
			name:      "printf_above_altscreen",
			altScreen: true,
			cmds:      sequenceMsg{PrintfAbove("log: %d", 42)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			var in bytes.Buffer

			var m Model
			if test.altScreen {
				m = &testViewModel{testModel: &testModel{}, opts: testViewOpts{altScreen: true}}
			} else {
				m = &testModel{}
			}

			p := NewProgram(m,
				WithWindowSize(80, 24),
				WithColorProfile(colorprofile.ANSI256),
				WithEnvironment([]string{"TERM=xterm-256color"}),
				WithInput(&in),
				WithOutput(&buf),
			)

			// Wait for the first render to flush so insertAbove observes an
			// established lastView.
			msgs := make(sequenceMsg, 0, len(test.cmds)+2)
			msgs = append(msgs, func() Msg { time.Sleep(20 * time.Millisecond); return nil })
			msgs = append(msgs, test.cmds...)
			msgs = append(msgs, Quit)
			go p.Send(msgs)

			if _, err := p.Run(); err != nil {
				t.Fatal(err)
			}
			golden.RequireEqual(t, buf.Bytes())
		})
	}
}

func TestPrintAboveCmds(t *testing.T) {
	tests := []struct {
		name    string
		cmd     Cmd
		body    string
		persist bool
	}{
		{name: "PrintlnAbove", cmd: PrintlnAbove("hello"), body: "hello", persist: true},
		{name: "PrintfAbove", cmd: PrintfAbove("val=%d", 7), body: "val=7", persist: true},
		{name: "Println", cmd: Println("regular"), body: "regular", persist: false},
		{name: "Printf", cmd: Printf("fmt=%s", "x"), body: "fmt=x", persist: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			msg := test.cmd()
			plm, ok := msg.(printLineMessage)
			if !ok {
				t.Fatalf("%s returned %T, want printLineMessage", test.name, msg)
			}
			if plm.persistOnAltScreen != test.persist {
				t.Errorf("%s: persistOnAltScreen = %v, want %v", test.name, plm.persistOnAltScreen, test.persist)
			}
			if plm.messageBody != test.body {
				t.Errorf("%s: messageBody = %q, want %q", test.name, plm.messageBody, test.body)
			}
		})
	}
}
