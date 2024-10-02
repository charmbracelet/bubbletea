package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
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
	username = iota
	pswd
	confirmPswd
	email
	signature
)

const (
	hotPink   = lipgloss.Color("#FF06B7")
	darkGray  = lipgloss.Color("#767676")
	candyRed  = lipgloss.Color("#FF0800")
	spearmint = lipgloss.Color("#45B08C")
)

var (
	inputStyle     = lipgloss.NewStyle().Foreground(hotPink)
	submitStyle    = lipgloss.NewStyle().Foreground(darkGray)
	errStyle       = lipgloss.NewStyle().Foreground(candyRed)
	signatureStyle = lipgloss.NewStyle().Foreground(spearmint).BorderStyle(lipgloss.RoundedBorder())
	successStyle   = lipgloss.NewStyle().Foreground(spearmint)
)

type model struct {
	inputs    []textinput.Model
	attempted bool
	focused   int
	err       error
}

// Validator functions to ensure valid input
func usernameValidator(s string) error {
	if s == "" {
		return errors.New("username must not be empty")
	}
	err := validateExistingUsername(s)
	if err != nil {
		return err
	}
	return nil
}

func pswdValidator(s string) error {
	var err error
	if s == "" {
		err = errors.Join(err, errors.New("password must not be empty"))
	}
	if len(s) < 8 {
		err = errors.Join(err, errors.New("password must be more than 8 characters"))
	}
	if lowerErr := validateLower(s); lowerErr != nil {
		err = errors.Join(err, lowerErr)
	}
	if upperErr := validateUpper(s); upperErr != nil {
		err = errors.Join(err, errors.New("password must have at least one uppercase letter"))
	}
	if numberErr := validateNumber(s); numberErr != nil {
		err = errors.Join(err, errors.New("password must have at least one number"))
	}
	if specialErr := validateSpecial(s); specialErr != nil {
		err = errors.Join(err, errors.New("password must have at least one special character"))
	}

	if err != nil {
		return err
	}
	return nil
}

func confirmPswdValidator(c string, p string) error {
	if c != p {
		return errors.New("passwords do not match")
	}
	return nil
}

func emailValidator(s string) error {
	// Official email address standard defined in RFC 5322
	const emailRegexRFC5322 = `(?i)^(?:[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)])$`

	re, err := regexp.Compile(emailRegexRFC5322)
	if err != nil {
		return errors.New("failed to validate email")
	}

	if !re.MatchString(s) {
		return errors.New("email is invalid")
	}

	return nil
}

func signatureValidator(s string) error {
	if s == "" {
		return errors.New("signature missing")
	}
	return nil
}

func initialModel() model {
	var inputs []textinput.Model = make([]textinput.Model, 5)
	inputs[username] = textinput.New()
	inputs[username].Placeholder = "Enter username"
	inputs[username].Focus()
	inputs[username].CharLimit = 20
	inputs[username].Width = 20
	inputs[username].Prompt = ""
	inputs[username].Validate = usernameValidator

	inputs[pswd] = textinput.New()
	inputs[pswd].Placeholder = "********"
	inputs[pswd].CharLimit = 50
	inputs[pswd].Width = 50
	inputs[pswd].Prompt = ""
	inputs[pswd].Validate = pswdValidator
	inputs[pswd].EchoMode = textinput.EchoPassword
	inputs[pswd].EchoCharacter = '*'

	inputs[confirmPswd] = textinput.New()
	inputs[confirmPswd].Placeholder = "********"
	inputs[confirmPswd].CharLimit = 50
	inputs[confirmPswd].Width = 50
	inputs[confirmPswd].Prompt = ""
	inputs[confirmPswd].Validate = func(c string) error {
		return confirmPswdValidator(c, inputs[pswd].Value())
	}
	inputs[confirmPswd].EchoMode = textinput.EchoPassword
	inputs[confirmPswd].EchoCharacter = '*'

	inputs[email] = textinput.New()
	inputs[email].Placeholder = "Enter email"
	inputs[email].CharLimit = 50
	inputs[email].Width = 50
	inputs[email].Prompt = ""
	inputs[email].Validate = emailValidator

	inputs[signature] = textinput.New()
	inputs[signature].Placeholder = ""
	inputs[signature].CharLimit = 50
	inputs[signature].Width = 50
	inputs[signature].Prompt = ""
	inputs[signature].Validate = signatureValidator

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
	err := m.getErrs()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				m.attempted = true
				if err == nil {
					fmt.Printf("%s\n", successStyle.Width(8).Render("Success!"))
					return m, tea.Quit
				}
			}
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
	err := m.getErrs()

	output := fmt.Sprintf(
		`
%s
%s

%s
%s

%s
%s

%s
%s


%s
%s

`,
		inputStyle.Width(30).Render("Username"),
		m.inputs[username].View(),
		inputStyle.Width(50).Render("Password"),
		m.inputs[pswd].View(),
		inputStyle.Width(50).Render("Confirm Password"),
		m.inputs[confirmPswd].View(),
		inputStyle.Width(50).Render("Email"),
		m.inputs[email].View(),
		inputStyle.Width(50).Render("Sign here"),
		signatureStyle.Width(50).Render(m.inputs[signature].View()),
	)

	// must attempt full submission at least once before showing errors
	if err != nil && m.attempted {
		output += fmt.Sprintf("%s\n", errStyle.Render(err.Error()))
	}

	output += fmt.Sprintf("%s\n", submitStyle.Render("Submit ->"))

	return output
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

func validateExistingUsername(s string) error {
	usernames := []string{"red01", "black07", "green89", "yellow29"}
	if slices.Contains(usernames, s) {
		return errors.New("username already exists")
	}
	return nil
}

func validateNumber(s string) error {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return nil
		}
	}

	return errors.New("password must have at least one number")
}

func validateUpper(s string) error {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return nil
		}
	}
	return errors.New("password must have at least one uppercase letter")
}

func validateLower(s string) error {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return nil
		}
	}
	return errors.New("password must have at least one lowercase letter")
}

func validateSpecial(s string) error {
	special := "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	for _, r := range s {
		if strings.Contains(special, string(r)) {
			return nil
		}
	}

	return errors.New("password must have at least one special character")
}

func (m model) getErrs() error {
	var err error
	for _, i := range m.inputs {
		if i.Err != nil {
			err = errors.Join(err, i.Err)
		}
	}

	if err != nil {
		return err
	}
	return nil
}
