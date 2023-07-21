package filepicker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"
	tea "github.com/rprtr258/bubbletea"
	"github.com/rprtr258/bubbletea/bubbles/key"
	"github.com/rprtr258/bubbletea/lipgloss"
)

var (
	lastID int
	idMtx  sync.Mutex
)

// Return the next ID we should use on the Model.
func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

// New returns a new filepicker model with default styling and key bindings.
func New() Model {
	return Model{
		id:               nextID(),
		CurrentDirectory: ".",
		Cursor:           ">",
		AllowedTypes:     []string{},
		selected:         0,
		ShowHidden:       false,
		DirAllowed:       false,
		FileAllowed:      true,
		AutoHeight:       true,
		Height:           0,
		max:              0,
		min:              0,
		selectedStack:    newStack(),
		minStack:         newStack(),
		maxStack:         newStack(),
		KeyMap:           DefaultKeyMap(),
		Styles:           DefaultStyles(),
	}
}

type errorMsg struct {
	err error
}

type readDirMsg struct {
	id      int
	entries []os.DirEntry
}

const (
	marginBottom  = 5
	fileSizeWidth = 8
	paddingLeft   = 2
)

// KeyMap defines key bindings for each user action.
type KeyMap struct {
	GoToTop  key.Binding
	GoToLast key.Binding
	Down     key.Binding
	Up       key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Back     key.Binding
	Open     key.Binding
	Select   key.Binding
}

// DefaultKeyMap defines the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		GoToTop:  key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "first")),
		GoToLast: key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "last")),
		Down:     key.NewBinding(key.WithKeys("j", "down", "ctrl+n"), key.WithHelp("j", "down")),
		Up:       key.NewBinding(key.WithKeys("k", "up", "ctrl+p"), key.WithHelp("k", "up")),
		PageUp:   key.NewBinding(key.WithKeys("K", "pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("J", "pgdown"), key.WithHelp("pgdown", "page down")),
		Back:     key.NewBinding(key.WithKeys("h", "backspace", "left", "esc"), key.WithHelp("h", "back")),
		Open:     key.NewBinding(key.WithKeys("l", "right", "enter"), key.WithHelp("l", "open")),
		Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	}
}

// Styles defines the possible customizations for styles in the file picker.
type Styles struct {
	DisabledCursor   lipgloss.Style
	Cursor           lipgloss.Style
	Symlink          lipgloss.Style
	Directory        lipgloss.Style
	File             lipgloss.Style
	DisabledFile     lipgloss.Style
	Permission       lipgloss.Style
	Selected         lipgloss.Style
	DisabledSelected lipgloss.Style
	FileSize         lipgloss.Style
	EmptyDirectory   lipgloss.Style
}

// DefaultStyles defines the default styling for the file picker.
func DefaultStyles() Styles {
	return DefaultStylesWithRenderer(lipgloss.DefaultRenderer())
}

// DefaultStylesWithRenderer defines the default styling for the file picker,
// with a given Lip Gloss renderer.
func DefaultStylesWithRenderer(r *lipgloss.Renderer) Styles {
	return Styles{
		DisabledCursor:   r.NewStyle().Foreground(lipgloss.Color("247")),
		Cursor:           r.NewStyle().Foreground(lipgloss.Color("212")),
		Symlink:          r.NewStyle().Foreground(lipgloss.Color("36")),
		Directory:        r.NewStyle().Foreground(lipgloss.Color("99")),
		File:             r.NewStyle(),
		DisabledFile:     r.NewStyle().Foreground(lipgloss.Color("243")),
		DisabledSelected: r.NewStyle().Foreground(lipgloss.Color("247")),
		Permission:       r.NewStyle().Foreground(lipgloss.Color("244")),
		Selected:         r.NewStyle().Foreground(lipgloss.Color("212")).Bold(true),
		FileSize:         r.NewStyle().Foreground(lipgloss.Color("240")).Width(fileSizeWidth).Align(lipgloss.Right),
		EmptyDirectory: r.NewStyle().
			Foreground(lipgloss.Color("240")).
			PaddingLeft(paddingLeft).
			SetString("Bummer. No Files Found."),
	}
}

// Model represents a file picker.
type Model struct {
	id int

	// Path is the path which the user has selected with the file picker.
	Path string

	// CurrentDirectory is the directory that the user is currently in.
	CurrentDirectory string

	// AllowedTypes specifies which file types the user may select.
	// If empty the user may select any file.
	AllowedTypes []string

	KeyMap      KeyMap
	files       []os.DirEntry
	ShowHidden  bool
	DirAllowed  bool
	FileAllowed bool

	FileSelected  string
	selected      int
	selectedStack stack

	min      int
	max      int
	maxStack stack
	minStack stack

	Height     int
	AutoHeight bool

	Cursor string
	Styles Styles
}

type stack struct {
	Push   func(int)
	Pop    func() int
	Length func() int
}

func newStack() stack {
	slice := make([]int, 0)
	return stack{
		Push: func(i int) {
			slice = append(slice, i)
		},
		Pop: func() int {
			res := slice[len(slice)-1]
			slice = slice[:len(slice)-1]
			return res
		},
		Length: func() int {
			return len(slice)
		},
	}
}

func (m Model) pushView() {
	m.minStack.Push(m.min)
	m.maxStack.Push(m.max)
	m.selectedStack.Push(m.selected)
}

func (m Model) popView() (int, int, int) {
	return m.selectedStack.Pop(), m.minStack.Pop(), m.maxStack.Pop()
}

func (m Model) readDir(path string, showHidden bool) tea.Cmd {
	return func() tea.Msg {
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			return errorMsg{err}
		}

		sort.Slice(dirEntries, func(i, j int) bool {
			if dirEntries[i].IsDir() == dirEntries[j].IsDir() {
				return dirEntries[i].Name() < dirEntries[j].Name()
			}
			return dirEntries[i].IsDir()
		})

		if showHidden {
			return readDirMsg{id: m.id, entries: dirEntries}
		}

		var sanitizedDirEntries []os.DirEntry
		for _, dirEntry := range dirEntries {
			isHidden, _ := IsHidden(dirEntry.Name())
			if isHidden {
				continue
			}
			sanitizedDirEntries = append(sanitizedDirEntries, dirEntry)
		}
		return readDirMsg{id: m.id, entries: sanitizedDirEntries}
	}
}

// Init initializes the file picker model.
func (m Model) Init() tea.Cmd {
	return m.readDir(m.CurrentDirectory, m.ShowHidden)
}

// Update handles user interactions within the file picker model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case readDirMsg:
		if msg.id != m.id {
			break
		}
		m.files = msg.entries
		m.max = m.Height - 1
	case tea.WindowSizeMsg:
		if m.AutoHeight {
			m.Height = msg.Height - marginBottom
		}
		m.max = m.Height - 1
	case tea.MsgKey:
		switch {
		case key.Matches(msg, m.KeyMap.GoToTop):
			m.selected = 0
			m.min = 0
			m.max = m.Height - 1
		case key.Matches(msg, m.KeyMap.GoToLast):
			m.selected = len(m.files) - 1
			m.min = len(m.files) - m.Height
			m.max = len(m.files) - 1
		case key.Matches(msg, m.KeyMap.Down):
			m.selected++
			if m.selected >= len(m.files) {
				m.selected = len(m.files) - 1
			}
			if m.selected > m.max {
				m.min++
				m.max++
			}
		case key.Matches(msg, m.KeyMap.Up):
			m.selected--
			if m.selected < 0 {
				m.selected = 0
			}
			if m.selected < m.min {
				m.min--
				m.max--
			}
		case key.Matches(msg, m.KeyMap.PageDown):
			m.selected += m.Height
			if m.selected >= len(m.files) {
				m.selected = len(m.files) - 1
			}
			m.min += m.Height
			m.max += m.Height

			if m.max >= len(m.files) {
				m.max = len(m.files) - 1
				m.min = m.max - m.Height
			}
		case key.Matches(msg, m.KeyMap.PageUp):
			m.selected -= m.Height
			if m.selected < 0 {
				m.selected = 0
			}
			m.min -= m.Height
			m.max -= m.Height

			if m.min < 0 {
				m.min = 0
				m.max = m.min + m.Height
			}
		case key.Matches(msg, m.KeyMap.Back):
			m.CurrentDirectory = filepath.Dir(m.CurrentDirectory)
			if m.selectedStack.Length() > 0 {
				m.selected, m.min, m.max = m.popView()
			} else {
				m.selected = 0
				m.min = 0
				m.max = m.Height - 1
			}
			return m, m.readDir(m.CurrentDirectory, m.ShowHidden)
		case key.Matches(msg, m.KeyMap.Open):
			if len(m.files) == 0 {
				break
			}

			f := m.files[m.selected]
			info, err := f.Info()
			if err != nil {
				break
			}
			isSymlink := info.Mode()&os.ModeSymlink != 0
			isDir := f.IsDir()

			if isSymlink {
				symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, f.Name()))
				info, err := os.Stat(symlinkPath)
				if err != nil {
					break
				}
				if info.IsDir() {
					isDir = true
				}
			}

			if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) {
				if key.Matches(msg, m.KeyMap.Select) {
					// Select the current path as the selection
					m.Path = filepath.Join(m.CurrentDirectory, f.Name())
				}
			}

			if !isDir {
				break
			}

			m.CurrentDirectory = filepath.Join(m.CurrentDirectory, f.Name())
			m.pushView()
			m.selected = 0
			m.min = 0
			m.max = m.Height - 1
			return m, m.readDir(m.CurrentDirectory, m.ShowHidden)
		}
	}
	return m, nil
}

// View returns the view of the file picker.
func (m Model) View() string {
	if len(m.files) == 0 {
		return m.Styles.EmptyDirectory.String()
	}
	var s strings.Builder

	for i, f := range m.files {
		if i < m.min || i > m.max {
			continue
		}

		info, _ := f.Info()
		isSymlink := info.Mode()&os.ModeSymlink != 0
		size := humanize.Bytes(uint64(info.Size()))
		name := f.Name()

		var symlinkPath string
		if isSymlink {
			symlinkPath, _ = filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, name))
		}

		disabled := !m.canSelect(name) && !f.IsDir()

		if m.selected == i {
			selected := fmt.Sprintf(" %s %"+fmt.Sprint(m.Styles.FileSize.GetWidth())+"s %s", info.Mode().String(), size, name)
			if isSymlink {
				selected = fmt.Sprintf("%s → %s", selected, symlinkPath)
			}
			if disabled {
				s.WriteString(m.Styles.DisabledSelected.Render(m.Cursor) + m.Styles.DisabledSelected.Render(selected))
			} else {
				s.WriteString(m.Styles.Cursor.Render(m.Cursor) + m.Styles.Selected.Render(selected))
			}
			s.WriteRune('\n')
			continue
		}

		style := m.Styles.File
		switch {
		case f.IsDir():
			style = m.Styles.Directory
		case isSymlink:
			style = m.Styles.Symlink
		case disabled:
			style = m.Styles.DisabledFile
		}

		fileName := style.Render(name)
		if isSymlink {
			fileName = fmt.Sprintf("%s → %s", fileName, symlinkPath)
		}
		s.WriteString(fmt.Sprintf(
			"  %s %s %s",
			m.Styles.Permission.Render(info.Mode().String()),
			m.Styles.FileSize.Render(size),
			fileName,
		))
		s.WriteRune('\n')
	}

	return s.String()
}

// DidSelectFile returns whether a user has selected a file (on this msg).
func (m Model) DidSelectFile(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelectFile(msg)
	return didSelect && m.canSelect(path), path
}

// DidSelectDisabledFile returns whether a user tried to select a disabled file
// (on this msg). This is necessary only if you would like to warn the user that
// they tried to select a disabled file.
func (m Model) DidSelectDisabledFile(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelectFile(msg)
	if didSelect && !m.canSelect(path) {
		return true, path
	}
	return false, ""
}

func (m Model) didSelectFile(msg tea.Msg) (bool, string) {
	if len(m.files) == 0 {
		return false, ""
	}

	switch msg := msg.(type) {
	case tea.MsgKey:
		// If the msg does not match the Select keymap then this could not have been a selection.
		if !key.Matches(msg, m.KeyMap.Select) {
			return false, ""
		}

		// The key press was a selection, let's confirm whether the current file could
		// be selected or used for navigating deeper into the stack.
		f := m.files[m.selected]
		info, err := f.Info()
		if err != nil {
			return false, ""
		}
		isSymlink := info.Mode()&os.ModeSymlink != 0
		isDir := f.IsDir()

		if isSymlink {
			symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, f.Name()))
			info, err := os.Stat(symlinkPath)
			if err != nil {
				break
			}
			if info.IsDir() {
				isDir = true
			}
		}

		if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) && m.Path != "" {
			return true, m.Path
		}

		// If the msg was not a MsgKey, then the file could not have been selected this iteration.
		// Only a MsgKey can select a file.
	default:
		return false, ""
	}
	return false, ""
}

func (m Model) canSelect(file string) bool {
	if len(m.AllowedTypes) == 0 {
		return true
	}

	for _, ext := range m.AllowedTypes {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}
