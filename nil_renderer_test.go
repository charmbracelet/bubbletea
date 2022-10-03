package tea

import "testing"

func TestNilRenderer(t *testing.T) {
	r := nilRenderer{}
	r.start()
	r.stop()
	r.kill()
	r.write("a")
	r.repaint()
	r.enterAltScreen()
	if r.altScreen() {
		t.Errorf("altScreen should always return false")
	}
}
