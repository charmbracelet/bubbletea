package main

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/charmtone"
)

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

var (
	inputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(charmtone.Tang.Hex()))
	continueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(charmtone.Anchovy.Hex()))
	validStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(charmtone.Guac.Hex()))
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(charmtone.Cherry.Hex()))
)

type model struct {
	input   textinput.Model
	focused int
	err     error
}

// canFindBook returns whether the find button is to be pressed
func (m model) canFindBook() bool {
	return m.input.Err == nil && len(m.input.Value()) != 0
}

// Validator function to ensure valid input
func isbn13Validator(s string) error {
	// A valid ISBN looks like this:
	// 978-3-548-37257-0 or
	// 9783548372570 without any spaces

	// Remove dashes
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 13 {
		return fmt.Errorf("ISBN is of wrong length")
	}

	for _, c := range s {
		if !unicode.IsDigit(c) {
			return fmt.Errorf("ISBN contains invalid characters")
		}
	}

	gs1Prefix := s[:3]
	switch gs1Prefix {
	case "978", "979":
		break
	default:
		return fmt.Errorf("ISBN has invalid GS1 prefix")
	}

	// The last digit, the check digit,
	// must make the checksum a multiple of 10.
	// All digits are added up after being multiplied
	// by either 1 or 3 alternately.
	// So 9x1 + 7x3 + 8x1 + ... + 0 must be a multiple of 10.

	sum := 0
	for i, c := range s {
		// Convert rune to int
		n := int(c - '0')

		// Multiply the uneven indices by 3
		if i%2 != 0 {
			n *= 3
		}

		sum += n
	}

	if sum%10 != 0 {
		return fmt.Errorf("ISBN has invalid check digit")
	}

	return nil
}

func initialModel() model {
	input := textinput.New()
	input = textinput.New()
	input.Placeholder = "978-X-XXX-XXXXX-X"
	input.Focus()
	input.CharLimit = 17
	input.Width = 30
	input.Prompt = ""
	input.Validate = isbn13Validator

	return model{
		input:   input,
		focused: 0,
		err:     nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Enter is blocked when ISBN is invalid or empty
			if m.canFindBook() {
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	return m, cmd
}

func (m model) View() string {
	var continueText string
	if m.canFindBook() {
		continueText = continueStyle.Render("Find ->")
	}

	var errorText string
	if m.input.Value() != "" {
		if m.input.Err != nil {
			errorText = errStyle.Render(m.input.Err.Error())
		} else {
			errorText = validStyle.Render("Valid ISBN")
		}
	}

	return fmt.Sprintf(
		` Search book:
 %s
 %s
 %s
 %s
`,
		inputStyle.Width(30).Render("ISBN"),
		m.input.View(),
		errorText,
		continueText,
	) + "\n"
}
