package tea

import "testing"

func TestNilRenderer(t *testing.T) {
	r := nilRenderer{}
	r.Repaint()
	r.EnterAltScreen()
	if r.AltScreen() {
		t.Errorf("altScreen should always return false")
	}
	r.ExitAltScreen()
	r.ClearScreen()
	r.ShowCursor()
	r.HideCursor()
}
