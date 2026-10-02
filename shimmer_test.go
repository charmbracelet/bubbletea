package tea

import (
	"image/color"
	"testing"
)

func TestShimmerText_Empty(t *testing.T) {
	got := ShimmerText("", 0.5, color.White)
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestShimmerText_NoColor(t *testing.T) {
	got := ShimmerText("hello", 0.5, nil)
	if got != "hello" {
		t.Fatalf("expected plain text, got %q", got)
	}
}

func TestShimmerText_ClampsProgress(t *testing.T) {
	got := ShimmerText("hello", -0.5, color.White)
	if got == "" {
		t.Fatal("expected non-empty output for negative progress")
	}

	got = ShimmerText("hello", 1.5, color.White)
	if got == "" {
		t.Fatal("expected non-empty output for over-1 progress")
	}
}

func TestShimmerText_ContainsColorEscape(t *testing.T) {
	got := ShimmerText("hello world", 0.5, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	if got == "hello world" {
		t.Fatal("expected ANSI color codes in output")
	}
}

func TestShimmerText_NoColorAtEdges(t *testing.T) {
	// When shimmer center is at 0 (start), only first chars should be colored
	got := ShimmerText("hello world", 0.0, color.RGBA{R: 255, G: 0, B: 0, A: 255}, 2.0)
	if got == "hello world" {
		t.Fatal("expected ANSI color codes at progress 0.0")
	}

	// When shimmer center is at 1 (end), only last chars should be colored
	got = ShimmerText("hello world", 1.0, color.RGBA{R: 255, G: 0, B: 0, A: 255}, 2.0)
	if got == "hello world" {
		t.Fatal("expected ANSI color codes at progress 1.0")
	}
}

func TestShimmerState_Reset(t *testing.T) {
	s := NewShimmerState("test", 100, color.White)
	s.Tick(0.5)
	if s.progress != 0.5 {
		t.Fatalf("expected 0.5, got %f", s.progress)
	}
	s.Reset()
	if s.progress != 0.0 {
		t.Fatalf("expected 0.0 after reset, got %f", s.progress)
	}
}

func TestShimmerState_TickWraps(t *testing.T) {
	s := NewShimmerState("test", 100, color.White)
	// Simulate 40 ticks of 0.03 each = 1.2, should wrap to 0.2
	for i := 0; i < 40; i++ {
		s.Tick(0.03)
	}
	expected := 0.2
	const epsilon = 0.0001
	if s.progress < expected-epsilon || s.progress > expected+epsilon {
		t.Fatalf("expected ~%f after wrap, got %f", expected, s.progress)
	}
}

func TestShimmerState_SetText(t *testing.T) {
	s := NewShimmerState("hello", 100, color.White)
	s.SetText("world")
	if s.Text() != "world" {
		t.Fatalf("expected 'world', got %q", s.Text())
	}
}