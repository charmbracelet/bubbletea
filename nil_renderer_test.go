package tea

import "testing"

func TestNilRenderer(t *testing.T) {
	r := nilRenderer{}
	r.Start()
	r.Stop()
	r.Kill()
	r.Write("a")
	r.Repaint()
	r.SetAltScreen(true)
	if r.AltScreen() {
		t.Errorf("altScreen should always return false")
	}
}
