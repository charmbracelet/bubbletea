package tea

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/colorprofile"
)

type profileUpgradeModel struct{}

func (m profileUpgradeModel) Init() Cmd               { return nil }
func (m profileUpgradeModel) Update(Msg) (Model, Cmd) { return m, nil }
func (profileUpgradeModel) View() View {
	return NewView("\x1b[38;2;107;80;255mhello\x1b[m")
}

type profileUpgradeOutput struct {
	mu           sync.Mutex
	buf          bytes.Buffer
	first        chan struct{}
	upgraded     chan struct{}
	firstOnce    sync.Once
	upgradedOnce sync.Once
}

func (w *profileUpgradeOutput) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, _ := w.buf.Write(p)
	if strings.Contains(w.buf.String(), "hello") {
		w.firstOnce.Do(func() { close(w.first) })
	}
	if strings.Contains(w.buf.String(), "38;2;107;80;255") {
		w.upgradedOnce.Do(func() { close(w.upgraded) })
	}
	return n, nil
}

// A capability response can arrive after the first frame. The model need not
// change its RGB-styled content just because the renderer gained true color.
func TestProgramCapabilityRepaintsColorProfile(t *testing.T) {
	for _, capability := range []string{"RGB", "Tc"} {
		t.Run(capability, func(t *testing.T) {
			out := &profileUpgradeOutput{first: make(chan struct{}), upgraded: make(chan struct{})}
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			p := NewProgram(profileUpgradeModel{},
				WithInput(nil),
				WithOutput(out),
				WithoutSignals(),
				WithWindowSize(20, 4),
				WithEnvironment([]string{"TERM=xterm-256color", "CLICOLOR_FORCE=1"}),
				WithContext(ctx),
			)
			done := make(chan error, 1)
			go func() { _, err := p.Run(); done <- err }()
			defer func() { cancel(); <-done }()
			select {
			case <-out.first:
			case <-ctx.Done():
				t.Fatal("initial frame not rendered")
			}
			p.Send(CapabilityMsg{Content: capability})
			select {
			case <-out.upgraded:
			case <-ctx.Done():
				t.Fatal("capability response never repainted unchanged content in true color")
			}
		})
	}
}

func TestCursedRendererColorProfileChange(t *testing.T) {
	for _, alt := range []bool{false, true} {
		name := "inline"
		if alt {
			name = "altscreen"
		}
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 20, 4)
			r.setNoInput(true)
			r.setColorProfile(colorprofile.ANSI256)
			v := NewView("\x1b[38;2;107;80;255mhello\x1b[m")
			v.AltScreen = alt
			r.render(v)
			if err := r.flush(false); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), "hello") || strings.Contains(out.String(), "38;2;") {
				t.Fatalf("initial 256-color frame: %q", out.String())
			}
			out.Reset()
			r.render(v)
			if err := r.flush(false); err != nil {
				t.Fatal(err)
			}
			if out.Len() != 0 {
				t.Fatalf("unchanged frame wrote %q", out.String())
			}
			r.setColorProfile(colorprofile.TrueColor)
			r.render(v)
			if err := r.flush(false); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), "38;2;107;80;255") || !strings.Contains(out.String(), "hello") {
				t.Fatalf("profile upgrade did not repaint the unchanged view in true color: %q", out.String())
			}

			// Reporting the same profile must retain the unchanged-frame fast path.
			out.Reset()
			r.setColorProfile(colorprofile.TrueColor)
			r.render(v)
			if err := r.flush(false); err != nil {
				t.Fatal(err)
			}
			if out.Len() != 0 {
				t.Fatalf("unchanged profile wrote %q", out.String())
			}

			// Downgrades must also re-encode the existing cells, not just future text.
			r.setColorProfile(colorprofile.ANSI256)
			r.render(v)
			if err := r.flush(false); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), "hello") || !strings.Contains(out.String(), "38;5;") || strings.Contains(out.String(), "38;2;") {
				t.Fatalf("profile downgrade did not repaint in 256 colors: %q", out.String())
			}
			out.Reset()
			r.render(v)
			if err := r.flush(false); err != nil {
				t.Fatal(err)
			}
			if out.Len() != 0 {
				t.Fatalf("unchanged downgraded frame wrote %q", out.String())
			}
		})
	}
}
