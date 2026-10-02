package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestRemoveFocusedEditor(t *testing.T) {
	m := newModel()
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = next.(model)
	next, _ = m.Update(tea.KeyPressMsg{Code: 'w', Mod: tea.ModCtrl})
	m = next.(model)
	next, _ = m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	m = next.(model)
	if got := m.inputs[0].Value(); got != "x" {
		t.Fatalf("surviving editor value = %q, want x", got)
	}
}
