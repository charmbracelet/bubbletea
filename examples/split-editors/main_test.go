package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestRemoveEditorPreservesFocus(t *testing.T) {
	for _, test := range []struct {
		name  string
		count int
		focus int
	}{
		{name: "two_first", count: 2, focus: 0},
		{name: "two_last", count: 2, focus: 1},
		{name: "three_first", count: 3, focus: 0},
		{name: "three_middle", count: 3, focus: 1},
		{name: "three_last", count: 3, focus: 2},
		{name: "six_last", count: 6, focus: 5},
	} {
		t.Run(test.name, func(t *testing.T) {
			m := newModel()
			update := func(msg tea.Msg) {
				next, _ := m.Update(msg)
				m = next.(model)
			}
			update(tea.WindowSizeMsg{Width: 120, Height: 30})
			for len(m.inputs) < test.count {
				update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
			}
			for range test.focus {
				update(tea.KeyPressMsg{Code: tea.KeyTab})
			}
			wantFocus := min(test.focus, test.count-2)
			m.inputs[wantFocus].SetValue("kept ")
			update(tea.KeyPressMsg{Code: 'w', Mod: tea.ModCtrl})
			if len(m.inputs) != test.count-1 || m.focus != wantFocus {
				t.Fatalf("after removal: count=%d focus=%d, want count=%d focus=%d", len(m.inputs), m.focus, test.count-1, wantFocus)
			}
			for i := range m.inputs {
				if got, want := m.inputs[i].Focused(), i == wantFocus; got != want {
					t.Errorf("editor %d focused=%t, want %t", i, got, want)
				}
			}
			update(tea.KeyPressMsg{Code: 'x', Text: "x"})
			if got := m.inputs[wantFocus].Value(); got != "kept x" {
				t.Errorf("surviving editor value=%q, want %q", got, "kept x")
			}
		})
	}
}
