package tea

import "testing"

func TestNilRenderer(t *testing.T) {
	r := NilRenderer{}
	r.Repaint()
	r.SetMode(altScreenMode, true)
	if r.Mode(altScreenMode) {
		t.Errorf("altScreen should always return false")
	}
	r.SetMode(altScreenMode, false)
	r.ClearScreen()
	r.SetMode(hideCursor, false)
	r.SetMode(hideCursor, true)
}
