package main

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestCommandsModel(t *testing.T) {
	m := model{}

	// Verify Init returns checkServer command
	if m.Init() == nil {
		t.Fatal("expected non-nil Init command")
	}

	// Update with statusMsg
	res, cmd := m.Update(statusMsg(200))
	m = res.(model)
	if m.status != 200 {
		t.Fatalf("expected status 200, got %d", m.status)
	}
	if cmd == nil {
		t.Fatal("expected quit command on statusMsg")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", cmd())
	}

	// View with success status
	view := m.View()
	if !strings.Contains(view.Content, "200 OK!") {
		t.Fatalf("expected status text in view content, got %q", view.Content)
	}

	// Update with errMsg
	testErr := errors.New("connection failed")
	res, cmd = m.Update(errMsg{err: testErr})
	m = res.(model)
	if m.err == nil || m.err.Error() != "connection failed" {
		t.Fatalf("expected error message, got %v", m.err)
	}
	if cmd == nil {
		t.Fatal("expected quit command on errMsg")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", cmd())
	}

	// View with error
	view = m.View()
	if !strings.Contains(view.Content, "We had some trouble: connection failed") {
		t.Fatalf("expected error description in view content, got %q", view.Content)
	}

	// Key Ctrl+C
	_, cmd = m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("expected quit command on Ctrl+C")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", cmd())
	}
}
