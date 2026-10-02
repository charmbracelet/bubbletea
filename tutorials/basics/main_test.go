package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestBasicsModel(t *testing.T) {
	m := initialModel()
	if len(m.choices) == 0 {
		t.Fatal("expected non-empty initial choices")
	}
	if m.cursor != 0 {
		t.Fatalf("expected initial cursor 0, got %d", m.cursor)
	}
	if m.Init() != nil {
		t.Fatal("expected nil Init command")
	}

	// Move cursor down
	res, _ := m.Update(tea.KeyPressMsg{Text: "down"})
	m = res.(model)
	if m.cursor != 1 {
		t.Fatalf("expected cursor 1 after down, got %d", m.cursor)
	}

	// Toggle selection on item 1
	res, _ = m.Update(tea.KeyPressMsg{Text: "enter"})
	m = res.(model)
	if _, ok := m.selected[1]; !ok {
		t.Fatal("expected item 1 to be selected")
	}

	// Move cursor up
	res, _ = m.Update(tea.KeyPressMsg{Text: "up"})
	m = res.(model)
	if m.cursor != 0 {
		t.Fatalf("expected cursor 0 after up, got %d", m.cursor)
	}

	// Toggle selection off on item 1
	res, _ = m.Update(tea.KeyPressMsg{Text: "down"})
	m = res.(model)
	res, _ = m.Update(tea.KeyPressMsg{Text: "enter"})
	m = res.(model)
	if _, ok := m.selected[1]; ok {
		t.Fatal("expected item 1 to be deselected")
	}

	// Quit command
	_, cmd := m.Update(tea.KeyPressMsg{Text: "q"})
	if cmd == nil {
		t.Fatal("expected non-nil quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg from quit command, got %T", cmd())
	}

	// View output
	view := m.View()
	if !strings.Contains(view.Content, "What should we buy at the market?") {
		t.Fatalf("expected title in view content, got %q", view.Content)
	}
	if view.WindowTitle != "Grocery List" {
		t.Fatalf("expected window title %q, got %q", "Grocery List", view.WindowTitle)
	}
}
