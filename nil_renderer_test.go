package tea

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilRenderer(t *testing.T) {
	r := nilRenderer{}
	r.start()
	r.stop()
	r.kill()
	r.Write("a")
	r.repaint()
	r.enterAltScreen()
	assert.False(t, r.altScreen(), "altScreen should always return false")
	r.exitAltScreen()
	r.clearScreen()
	r.showCursor()
	r.hideCursor()
	r.enableMouseCellMotion()
	r.disableMouseCellMotion()
	r.enableMouseAllMotion()
	r.disableMouseAllMotion()
}
