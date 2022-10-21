package tea

import (
	"testing"

	"github.com/muesli/termenv"
)

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
	r.exitAltScreen()
	r.clearScreen()
	r.showCursor()
	r.hideCursor()
	r.enableMouseCellMotion()
	r.disableMouseCellMotion()
	r.enableMouseAllMotion()
	r.disableMouseAllMotion()
	r.setBackgroundColor(termenv.RGBColor("#ffffff"))
	r.resetBackgroundColor()
}
