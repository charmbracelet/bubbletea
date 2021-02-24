package main

// An example of how to pipe in data to a Bubble Tea application. It's actually
// more of a proof that Bubble Tea will automatically listen for keystrokes
// when input is not a TTY, such as when data is piped or redirected in.
//
// In the case of this example we're listing for a single keystroke used to
// exit the program.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0 {
		fmt.Println("Try piping in some text.")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	var b strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		_, err = b.WriteRune(r)
		if err != nil {
			fmt.Println("Error getting input:", err)
			os.Exit(1)
		}
	}

	model := newModel(strings.TrimSpace(b.String()))

	if err := tea.NewProgram(model).Start(); err != nil {
		fmt.Println("Couldn't start program:", err)
		os.Exit(1)
	}
}

type model struct {
	userInput textinput.Model
}

func newModel(initialValue string) (m model) {
	i := textinput.NewModel()
	i.Prompt = ""
	i.CursorColor = "63"
	i.Width = 48
	i.SetValue(initialValue)
	i.CursorEnd()
	i.Focus()

	m.userInput = i
	return
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.userInput, cmd = m.userInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"\nYou piped in: %s\n\nPress ^C to exit",
		m.userInput.View(),
	)
}
