package main

// This example demonstrates how to use a custom Lip Gloss renderer with Wish,
// a package for building custom SSH servers.
//
// The big advantage to using custom renderers here is that we can accurately
// detect the background color and color profile for each client and render
// against that accordingly.
//
// For details on wish see: https://github.com/charmbracelet/wish/

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/kr/pty"
	"github.com/muesli/termenv"
	"github.com/rprtr258/bubbletea/lipgloss"
)

// Available styles.
type styles struct {
	bold          lipgloss.Style
	faint         lipgloss.Style
	italic        lipgloss.Style
	underline     lipgloss.Style
	strikethrough lipgloss.Style
	red           lipgloss.Style
	green         lipgloss.Style
	yellow        lipgloss.Style
	blue          lipgloss.Style
	magenta       lipgloss.Style
	cyan          lipgloss.Style
	gray          lipgloss.Style
}

// Create new styles against a given renderer.
func makeStyles(r *lipgloss.Renderer) styles {
	return styles{
		bold:          r.NewStyle().SetString("bold").Bold(true),
		faint:         r.NewStyle().SetString("faint").Faint(true),
		italic:        r.NewStyle().SetString("italic").Italic(true),
		underline:     r.NewStyle().SetString("underline").Underline(true),
		strikethrough: r.NewStyle().SetString("strikethrough").Strikethrough(true),
		red:           r.NewStyle().SetString("red").Foreground(lipgloss.Color("#E88388")),
		green:         r.NewStyle().SetString("green").Foreground(lipgloss.Color("#A8CC8C")),
		yellow:        r.NewStyle().SetString("yellow").Foreground(lipgloss.Color("#DBAB79")),
		blue:          r.NewStyle().SetString("blue").Foreground(lipgloss.Color("#71BEF2")),
		magenta:       r.NewStyle().SetString("magenta").Foreground(lipgloss.Color("#D290E4")),
		cyan:          r.NewStyle().SetString("cyan").Foreground(lipgloss.Color("#66C2CD")),
		gray:          r.NewStyle().SetString("gray").Foreground(lipgloss.Color("#B9BFCA")),
	}
}

// Bridge Wish and Termenv so we can query for a user's terminal capabilities.
type sshOutput struct {
	ssh.Session
	tty *os.File
}

func (s *sshOutput) Write(p []byte) (int, error) {
	return s.Session.Write(p)
}

func (s *sshOutput) Read(p []byte) (int, error) {
	return s.Session.Read(p)
}

func (s *sshOutput) Fd() uintptr {
	return s.tty.Fd()
}

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

// Create a termenv.Output from the session.
func outputFromSession(sess ssh.Session) *termenv.Output {
	sshPty, _, _ := sess.Pty()
	_, tty, err := pty.Open()
	if err != nil {
		log.Fatal(err)
	}
	o := &sshOutput{
		Session: sess,
		tty:     tty,
	}
	environ := sess.Environ()
	environ = append(environ, fmt.Sprintf("TERM=%s", sshPty.Term))
	e := &sshEnviron{environ: environ}
	// We need to use unsafe mode here because the ssh session is not running
	// locally and we already know that the session is a TTY.
	return termenv.NewOutput(o, termenv.WithUnsafe(), termenv.WithEnvironment(e))
}

// Handle SSH requests.
func handler(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		// Get client's output.
		clientOutput := outputFromSession(sess)

		pty, _, active := sess.Pty()
		if !active {
			next(sess)
			return
		}
		width := pty.Window.Width

		// Initialize new renderer for the client.
		renderer := lipgloss.NewRenderer(sess)
		renderer.SetOutput(clientOutput)

		// Initialize new styles against the renderer.
		styles := makeStyles(renderer)

		str := strings.Builder{}

		fmt.Fprintf(&str, "\n\n%s %s %s %s %s",
			styles.bold,
			styles.faint,
			styles.italic,
			styles.underline,
			styles.strikethrough,
		)

		fmt.Fprintf(&str, "\n%s %s %s %s %s %s %s",
			styles.red,
			styles.green,
			styles.yellow,
			styles.blue,
			styles.magenta,
			styles.cyan,
			styles.gray,
		)

		fmt.Fprintf(&str, "\n%s %s %s %s %s %s %s\n\n",
			styles.red,
			styles.green,
			styles.yellow,
			styles.blue,
			styles.magenta,
			styles.cyan,
			styles.gray,
		)

		fmt.Fprintf(&str, "%s %t %s\n\n", styles.bold.Copy().UnsetString().Render("Has dark background?"),
			renderer.HasDarkBackground(),
			renderer.Output().BackgroundColor())

		block := renderer.Place(width,
			lipgloss.Height(str.String()), lipgloss.Center, lipgloss.Center, str.String(),
			lipgloss.WithWhitespaceChars("/"),
			lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{Light: "250", Dark: "236"}),
		)

		// Render to client.
		wish.WriteString(sess, block)

		next(sess)
	}
}

func main() {
	port := 3456
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf(":%d", port)),
		wish.WithHostKeyPath("ssh_example"),
		wish.WithMiddleware(handler, lm.Middleware()),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SSH server listening on port %d", port)
	log.Printf("To connect from your local machine run: ssh localhost -p %d", port)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
