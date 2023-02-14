package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

const (
	ccn = iota
	exp
	cvv
)

var (
	submitting bool
)

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
	errorRed = lipgloss.Color("#FF2222")
)

var (
	inputStyle    = lipgloss.NewStyle().Foreground(hotPink)
	errorStyle    = lipgloss.NewStyle().Foreground(errorRed)
	continueStyle = lipgloss.NewStyle().Foreground(darkGray)
)

type model struct {
	inputs  []textinput.Model
	focused int
	err     error
}

// Validator functions to ensure valid input
// Note: Please don't mix runtime errors with your users' errors
func ccnValidator(s string) error {
	// Credit Card Number should a string less than 20 digits
	// It should include 16 integers and 3 spaces

	if len(s) == 0 {
		if submitting {
			return fmt.Errorf("CCN cannot be blank")
		}
		return nil
	}

	// Since typing past 19 is intuitively disabled, we don't *really* need a msg
	// This is brings ccnValidator in line with the other validators
	if len(s) > 16+3 {
		return fmt.Errorf("")
	}

	// The last digit should be a number unless it is a multiple of 4 in which
	// case it should be a space
	if len(s)%5 == 0 && s[len(s)-1] != ' ' {
		return fmt.Errorf("CCN must separate groups with spaces")
	}
	if len(s)%5 != 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		return fmt.Errorf("CCN must be numbers seperated by spaces")
	}

	// The remaining digits should be integers
	c := strings.ReplaceAll(s, " ", "")
	_, err := strconv.ParseInt(c, 10, 64)

	return err
}

func expValidator(s string) error {
	// The 3 character should be a slash (/)
	// The rest thould be numbers
	if len(s) == 0 {
		if submitting {
			return fmt.Errorf("EXP cannot be blank")
		}
		return nil
	}

	if len(s) >= 2 && s[1] == '/' {
		return fmt.Errorf("EXP must be two, two-digit numbers seperated by a slash (/)")
	}

	e := strings.ReplaceAll(s, "/", "")
	_, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		return fmt.Errorf("EXP must be two, two-digit numbers seperated by a slash (/)")
	}

	// There should be only one slash and it should be in the 2nd index (3rd character)
	if len(s) >= 3 && (strings.Index(s, "/") != 2 || strings.LastIndex(s, "/") != 2) {
		return fmt.Errorf("EXP month and year must be separated by a /")
	}

	return nil
}

func cvvValidator(s string) error {
	// The CVV should be a number of 3 digits
	if len(s) == 0 {
		if submitting {
			return fmt.Errorf("CVV cannot be blank")
		}
		return nil

	}
	// All we need to do is check that it is a number
	_, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("CVV must be a number")
	}
	return nil
}

func initialModel() model {
	var inputs []textinput.Model = make([]textinput.Model, 3)
	inputs[ccn] = textinput.New()
	inputs[ccn].Placeholder = "4505 **** **** 1234"
	inputs[ccn].Focus()
	inputs[ccn].CharLimit = 20
	inputs[ccn].Width = 30
	inputs[ccn].Prompt = ""
	inputs[ccn].Validate = ccnValidator

	inputs[exp] = textinput.New()
	inputs[exp].Placeholder = "MM/YY "
	inputs[exp].CharLimit = 5
	inputs[exp].Width = 5
	inputs[exp].Prompt = ""
	inputs[exp].Validate = expValidator

	inputs[cvv] = textinput.New()
	inputs[cvv].Placeholder = "XXX"
	inputs[cvv].CharLimit = 3
	inputs[cvv].Width = 5
	inputs[cvv].Prompt = ""
	inputs[cvv].Validate = cvvValidator

	return model{
		inputs:  inputs,
		focused: 0,
		err:     nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		submitting = false
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				return m, nil
			}
			submitting = true
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var err error
	errorString := ""
	err = m.inputs[m.focused].Err
	if err == nil {
		for i, input := range m.inputs {
			if i > m.focused {
				break
			}
			if input.Err != nil {
				err = input.Err
				break
			}
		}
	}
	if err != nil {
		errorString = err.Error()
	}

	return fmt.Sprintf(
		` Total: $21.50:

 %s
 %s

 %s  %s
 %s  %s
 %s
 %s
`,
		inputStyle.Width(30).Render("Card Number"),
		m.inputs[ccn].View(),
		inputStyle.Width(6).Render("EXP"),
		inputStyle.Width(6).Render("CVV"),
		m.inputs[exp].View(),
		m.inputs[cvv].View(),
		errorStyle.Render(errorString),
		continueStyle.Render("Continue ->"),
	) + "\n"
}

// nextInput focuses the next input field
func (m *model) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *model) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}
