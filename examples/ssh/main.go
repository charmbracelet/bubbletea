package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"
)

type sshEnviron struct {
	environ []string
}

func (s *sshEnviron) Getenv(key string) string {
	for _, v := range s.environ {
		if strings.HasPrefix(v, key+"=") {
			return v[len(key)+1:]
		}
	}
	return ""
}

func (s *sshEnviron) Environ() []string {
	return s.environ
}

type Styles struct {
	Bold      lipgloss.Style
	Faint     lipgloss.Style
	Italic    lipgloss.Style
	Underline lipgloss.Style
	Crossout  lipgloss.Style

	Red     lipgloss.Style
	Green   lipgloss.Style
	Yellow  lipgloss.Style
	Blue    lipgloss.Style
	Magenta lipgloss.Style
	Cyan    lipgloss.Style
	Gray    lipgloss.Style

	RedBg     lipgloss.Style
	GreenBg   lipgloss.Style
	YellowBg  lipgloss.Style
	BlueBg    lipgloss.Style
	MagentaBg lipgloss.Style
	CyanBg    lipgloss.Style
	GrayBg    lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Bold:      lipgloss.NewStyle().Bold(true),
		Faint:     lipgloss.NewStyle().Faint(true),
		Italic:    lipgloss.NewStyle().Italic(true),
		Underline: lipgloss.NewStyle().Underline(true),
		Crossout:  lipgloss.NewStyle().Strikethrough(true),
		Red:       lipgloss.NewStyle().Foreground(lipgloss.Color("#E88388")),
		Green:     lipgloss.NewStyle().Foreground(lipgloss.Color("#A8CC8C")),
		Yellow:    lipgloss.NewStyle().Foreground(lipgloss.Color("#DBAB79")),
		Blue:      lipgloss.NewStyle().Foreground(lipgloss.Color("#71BEF2")),
		Magenta:   lipgloss.NewStyle().Foreground(lipgloss.Color("#D290E4")),
		Cyan:      lipgloss.NewStyle().Foreground(lipgloss.Color("#66C2CD")),
		Gray:      lipgloss.NewStyle().Foreground(lipgloss.Color("#B9BFCA")),
		RedBg:     lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#E88388")),
		GreenBg:   lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#A8CC8C")),
		YellowBg:  lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#DBAB79")),
		BlueBg:    lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#71BEF2")),
		MagentaBg: lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#D290E4")),
		CyanBg:    lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#66C2CD")),
		GrayBg:    lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#B9BFCA")),
	}
}

func (s Styles) Renderer(r *lipgloss.Renderer) Styles {
	s.Bold = s.Bold.Copy().Renderer(r)
	s.Faint = s.Faint.Copy().Renderer(r)
	s.Italic = s.Italic.Copy().Renderer(r)
	s.Underline = s.Underline.Copy().Renderer(r)
	s.Crossout = s.Crossout.Copy().Renderer(r)
	s.Red = s.Red.Copy().Renderer(r)
	s.Green = s.Green.Copy().Renderer(r)
	s.Yellow = s.Yellow.Copy().Renderer(r)
	s.Blue = s.Blue.Copy().Renderer(r)
	s.Magenta = s.Magenta.Copy().Renderer(r)
	s.Cyan = s.Cyan.Copy().Renderer(r)
	s.Gray = s.Gray.Copy().Renderer(r)
	s.RedBg = s.RedBg.Copy().Renderer(r)
	s.GreenBg = s.GreenBg.Copy().Renderer(r)
	s.YellowBg = s.YellowBg.Copy().Renderer(r)
	s.BlueBg = s.BlueBg.Copy().Renderer(r)
	s.MagentaBg = s.MagentaBg.Copy().Renderer(r)
	s.CyanBg = s.CyanBg.Copy().Renderer(r)
	s.GrayBg = s.GrayBg.Copy().Renderer(r)

	return s
}

type model struct {
	session  ssh.Session
	renderer *lipgloss.Renderer

	styles Styles
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s := msg.String(); s == "ctrl+c" || s == "q" || s == "esc" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(m.styles.Bold.Render("bold"))
	b.WriteString(" ")
	b.WriteString(m.styles.Faint.Render("faint"))
	b.WriteString(" ")
	b.WriteString(m.styles.Italic.Render("italic"))
	b.WriteString(" ")
	b.WriteString(m.styles.Underline.Render("underline"))
	b.WriteString(" ")
	b.WriteString(m.styles.Crossout.Render("crossout"))
	b.WriteString("\n\n")

	b.WriteString(m.styles.Red.Render("red"))
	b.WriteString(" ")
	b.WriteString(m.styles.Green.Render("green"))
	b.WriteString(" ")
	b.WriteString(m.styles.Yellow.Render("yellow"))
	b.WriteString(" ")
	b.WriteString(m.styles.Blue.Render("blue"))
	b.WriteString(" ")
	b.WriteString(m.styles.Magenta.Render("magenta"))
	b.WriteString(" ")
	b.WriteString(m.styles.Cyan.Render("cyan"))
	b.WriteString(" ")
	b.WriteString(m.styles.Gray.Render("gray"))
	b.WriteString("\n\n")

	b.WriteString(m.styles.RedBg.Render("red"))
	b.WriteString(" ")
	b.WriteString(m.styles.GreenBg.Render("green"))
	b.WriteString(" ")
	b.WriteString(m.styles.YellowBg.Render("yellow"))
	b.WriteString(" ")
	b.WriteString(m.styles.BlueBg.Render("blue"))
	b.WriteString(" ")
	b.WriteString(m.styles.MagentaBg.Render("magenta"))
	b.WriteString(" ")
	b.WriteString(m.styles.CyanBg.Render("cyan"))
	b.WriteString(" ")
	b.WriteString(m.styles.GrayBg.Render("gray"))
	b.WriteString("\n\n")

	output := m.renderer.Output()
	b.WriteString(fmt.Sprintf("Has foreground color %s", output.ForegroundColor()))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Has background color %s", output.BackgroundColor()))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Has dark background?: %t", m.renderer.HasDarkBackground()))
	b.WriteString("\n\n")

	return b.String()
}

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(":2345"),
		wish.WithHostKeyPath("examples"),
		wish.WithMiddleware(
			bm.MiddlewareWithProgramHandler(func(s ssh.Session) *tea.Program {
				ptyReq, _, active := s.Pty()
				if !active {
					wish.Println(s, "not a pty")
					return nil
				}

				e := &sshEnviron{environ: append(s.Environ(), "TERM="+ptyReq.Term)}
				renderer := lipgloss.NewRenderer(s,
					// XXX: notice that order here is important since
					// termenv.WithColorCache depends on unsafe and environment
					// values.
					termenv.WithUnsafe(),
					termenv.WithEnvironment(e),
					termenv.WithColorCache(true),
				)

				m, opts := model{
					session:  s,
					renderer: renderer,
					styles:   DefaultStyles().Renderer(renderer),
				}, []tea.ProgramOption{
					tea.WithAltScreen(),
					tea.WithInput(s),
					tea.WithOutput(renderer.Output()),
				}

				return tea.NewProgram(m, opts...)
			}, termenv.ANSI256),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on %s", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
