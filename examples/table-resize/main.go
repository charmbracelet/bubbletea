package main

import (
	"fmt"
	"image/color"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
)

// Pokemon types.
const (
	None     = ""
	Bug      = "Bug"
	Electric = "Electric"
	Fire     = "Fire"
	Flying   = "Flying"
	Grass    = "Grass"
	Ground   = "Ground"
	Normal   = "Normal"
	Poison   = "Poison"
	Water    = "Water"
)

type model struct {
	table *table.Table
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.table = m.table.Width(msg.Width)
		m.table = m.table.Height(msg.Height)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
		}
	}
	return m, cmd
}

func (m model) View() tea.View {
	v := tea.NewView("\n" + m.table.String() + "\n")
	v.AltScreen = true
	return v
}

func main() {
	baseStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := baseStyle.Foreground(lipgloss.Color("252")).Bold(true)
	selectedStyle := baseStyle.Foreground(lipgloss.Color("#01BE85")).Background(lipgloss.Color("#00432F"))
	typeColors := map[string]color.Color{
		Bug:      lipgloss.Color("#D7FF87"),
		Electric: lipgloss.Color("#FDFF90"),
		Fire:     lipgloss.Color("#FF7698"),
		Flying:   lipgloss.Color("#FF87D7"),
		Grass:    lipgloss.Color("#75FBAB"),
		Ground:   lipgloss.Color("#FF875F"),
		Normal:   lipgloss.Color("#929292"),
		Poison:   lipgloss.Color("#7D5AFC"),
		Water:    lipgloss.Color("#00E2C7"),
	}
	dimTypeColors := map[string]color.Color{
		Bug:      lipgloss.Color("#97AD64"),
		Electric: lipgloss.Color("#FCFF5F"),
		Fire:     lipgloss.Color("#BA5F75"),
		Flying:   lipgloss.Color("#C97AB2"),
		Grass:    lipgloss.Color("#59B980"),
		Ground:   lipgloss.Color("#C77252"),
		Normal:   lipgloss.Color("#727272"),
		Poison:   lipgloss.Color("#634BD0"),
		Water:    lipgloss.Color("#439F8E"),
	}
	headers := []string{"#", "NAME", "TYPE 1", "TYPE 2", "JAPANESE", "OFFICIAL ROM."}
	rows := [][]string{
		{"1", "Bulbasaur", Grass, Poison, "フシギダネ", "Bulbasaur"},
		{"2", "Ivysaur", Grass, Poison, "フシギソウ", "Ivysaur"},
		{"3", "Venusaur", Grass, Poison, "フシギバナ", "Venusaur"},
		{"4", "Charmander", Fire, None, "ヒトカゲ", "Hitokage"},
		{"5", "Charmeleon", Fire, None, "リザード", "Lizardo"},
		{"6", "Charizard", Fire, Flying, "リザードン", "Lizardon"},
		{"7", "Squirtle", Water, None, "ゼニガメ", "Zenigame"},
		{"8", "Wartortle", Water, None, "カメール", "Kameil"},
		{"9", "Blastoise", Water, None, "カメックス", "Kamex"},
		{"10", "Caterpie", Bug, None, "キャタピー", "Caterpie"},
		{"11", "Metapod", Bug, None, "トランセル", "Trancell"},
		{"12", "Butterfree", Bug, Flying, "バタフリー", "Butterfree"},
		{"13", "Weedle", Bug, Poison, "ビードル", "Beedle"},
		{"14", "Kakuna", Bug, Poison, "コクーン", "Cocoon"},
		{"15", "Beedrill", Bug, Poison, "スピアー", "Spear"},
		{"16", "Pidgey", Normal, Flying, "ポッポ", "Poppo"},
		{"17", "Pidgeotto", Normal, Flying, "ピジョン", "Pigeon"},
		{"18", "Pidgeot", Normal, Flying, "ピジョット", "Pigeot"},
		{"19", "Rattata", Normal, None, "コラッタ", "Koratta"},
		{"20", "Raticate", Normal, None, "ラッタ", "Ratta"},
		{"21", "Spearow", Normal, Flying, "オニスズメ", "Onisuzume"},
		{"22", "Fearow", Normal, Flying, "オニドリル", "Onidrill"},
		{"23", "Ekans", Poison, None, "アーボ", "Arbo"},
		{"24", "Arbok", Poison, None, "アーボック", "Arbok"},
		{"25", "Pikachu", Electric, None, "ピカチュウ", "Pikachu"},
		{"26", "Raichu", Electric, None, "ライチュウ", "Raichu"},
		{"27", "Sandshrew", Ground, None, "サンド", "Sand"},
		{"28", "Sandslash", Ground, None, "サンドパン", "Sandpan"},
	}

	t := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}

			rowIndex := row - 1
			if rowIndex < 0 || rowIndex >= len(rows) {
				return baseStyle
			}

			if rows[rowIndex][1] == "Pikachu" {
				return selectedStyle
			}

			even := row%2 == 0

			switch col {
			case 2, 3: // Type 1 + 2
				c := typeColors
				if even {
					c = dimTypeColors
				}

				if col >= len(rows[rowIndex]) {
					return baseStyle
				}

				color, ok := c[rows[rowIndex][col]]
				if !ok {
					return baseStyle
				}
				return baseStyle.Foreground(color)
			}

			if even {
				return baseStyle.Foreground(lipgloss.Color("245"))
			}
			return baseStyle.Foreground(lipgloss.Color("252"))
		}).
		Border(lipgloss.ThickBorder())

	if _, err := tea.NewProgram(model{t}).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
