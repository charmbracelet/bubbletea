package tea

import (
	"testing"
)

func TestWithHardTabs(t *testing.T) {
	// Test that WithHardTabs sets the hardTabsOverride field
	p := &Program{}

	// Test enabling hard tabs
	opt := WithHardTabs(true)
	opt(p)
	if p.hardTabsOverride == nil || *p.hardTabsOverride != true {
		t.Error("WithHardTabs(true) did not set hardTabsOverride to true")
	}

	// Test disabling hard tabs
	opt = WithHardTabs(false)
	opt(p)
	if p.hardTabsOverride == nil || *p.hardTabsOverride != false {
		t.Error("WithHardTabs(false) did not set hardTabsOverride to false")
	}
}

func TestWithBackspace(t *testing.T) {
	// Test that WithBackspace sets the backspaceOverride field
	p := &Program{}

	// Test enabling backspace
	opt := WithBackspace(true)
	opt(p)
	if p.backspaceOverride == nil || *p.backspaceOverride != true {
		t.Error("WithBackspace(true) did not set backspaceOverride to true")
	}

	// Test disabling backspace
	opt = WithBackspace(false)
	opt(p)
	if p.backspaceOverride == nil || *p.backspaceOverride != false {
		t.Error("WithBackspace(false) did not set backspaceOverride to false")
	}
}

func TestWithHardTabsAndBackspaceCombined(t *testing.T) {
	// Test that both options can be used together
	p := &Program{}

	WithHardTabs(true)(p)
	WithBackspace(true)(p)

	if p.hardTabsOverride == nil || *p.hardTabsOverride != true {
		t.Error("WithHardTabs(true) did not persist after WithBackspace")
	}
	if p.backspaceOverride == nil || *p.backspaceOverride != true {
		t.Error("WithBackspace(true) did not persist after WithHardTabs")
	}
}

func TestWithHardTabsNilPointer(t *testing.T) {
	// Test that the function handles nil pointer correctly
	p := &Program{}

	// Before calling the option, hardTabsOverride should be nil
	if p.hardTabsOverride != nil {
		t.Error("hardTabsOverride should be nil before calling WithHardTabs")
	}

	// After calling, it should be set
	WithHardTabs(true)(p)
	if p.hardTabsOverride == nil {
		t.Error("hardTabsOverride should not be nil after calling WithHardTabs")
	}
}
