package tea

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// linesMatchForShift reports whether two lines represent the same content for
// the purpose of scroll-shift detection. It first tries exact equality (zero
// overhead when lines are identical), then falls back to a prefix comparison
// that ignores up to maxTrailingDiff bytes at the end. This tolerates styled
// trailing decorations such as scrollbar characters whose ANSI escapes cause
// the suffix to differ between frames even when the content is the same.
func linesMatchForShift(a, b string) bool {
	if a == b {
		return true
	}
	n := min(len(a), len(b))
	const maxTrailingDiff = 40
	if n <= maxTrailingDiff {
		return false
	}
	compareLen := n - maxTrailingDiff
	return a[:compareLen] == b[:compareLen]
}

// detectContentShift compares old and new content lines to detect a vertical
// scroll within the frame. Returns a positive shift for scroll-up (content
// moved up, new lines appear at bottom) and negative for scroll-down. Returns
// 0 if no shift is detected or dimensions changed.
//
// regionStart is the index of the first line in the scrolling region (after
// stripping identical leading chrome). matchCount is the number of consecutive
// lines within that region that matched the shifted pattern.
//
// The algorithm strips identical leading and trailing lines (non-scrolling
// chrome such as padding, status bars, input areas) before detecting shift.
// This allows apps with fixed chrome at both top and bottom of the frame to
// benefit from scroll optimisation.
func detectContentShift(oldLines, newLines []string) (shift, regionStart, matchCount int) {
	if len(oldLines) != len(newLines) || len(oldLines) < 4 {
		return 0, 0, 0
	}
	height := len(oldLines)

	top := 0
	for top < height && oldLines[top] == newLines[top] {
		top++
	}
	bot := height
	for bot > top && oldLines[bot-1] == newLines[bot-1] {
		bot--
	}

	regionHeight := bot - top
	if regionHeight < 4 {
		return 0, 0, 0
	}

	maxShift := min(10, regionHeight/2)
	minMatch := max(4, regionHeight/4)
	old := oldLines[top:bot]
	new_ := newLines[top:bot]

	for shift := 1; shift <= maxShift; shift++ {
		count := 0
		for i := 0; i+shift < regionHeight; i++ {
			if !linesMatchForShift(old[i+shift], new_[i]) {
				break
			}
			count++
		}
		if count >= minMatch {
			return shift, top, count
		}
	}

	for shift := 1; shift <= maxShift; shift++ {
		count := 0
		for i := 0; i+shift < regionHeight; i++ {
			if !linesMatchForShift(old[i], new_[i+shift]) {
				break
			}
			count++
		}
		if count >= minMatch {
			return -shift, top, count
		}
	}

	return 0, 0, 0
}

// shiftCellbufRegion shifts lines [regionStart, regionEnd) of the cell buffer
// by n positions in-place. Lines outside this range are left untouched.
// Positive n = scroll up; negative n = scroll down.
// After shifting, ALL lines get a nil Touched reset so updateHashmap and
// transformLine skip them; DrawOver sets Touched only on genuinely changed lines.
func shiftCellbufRegion(buf *uv.ScreenBuffer, regionStart, regionEnd, n int) {
	height := buf.Height()
	if regionEnd > height {
		regionEnd = height
	}
	regionLen := regionEnd - regionStart
	if regionLen < 1 {
		return
	}
	region := buf.Lines[regionStart:regionEnd]
	if n < 0 {
		n = -n
		if n >= regionLen {
			return
		}
		var saved [10]uv.Line
		copy(saved[:n], region[regionLen-n:regionLen])
		copy(region[n:regionLen], region[:regionLen-n])
		for i := 0; i < n; i++ {
			clearLine(saved[i])
			region[i] = saved[i]
		}
	} else {
		if n >= regionLen {
			return
		}
		var saved [10]uv.Line
		copy(saved[:n], region[:n])
		copy(region[:regionLen-n], region[n:regionLen])
		for i := 0; i < n; i++ {
			clearLine(saved[i])
			region[regionLen-n+i] = saved[i]
		}
	}

	for i := range buf.Touched {
		buf.Touched[i] = nil
	}
}

// clearCellbufLine sets all cells in line i to EmptyCell and resets its
// Touched sentinel so DrawOver writes the line fresh.
func clearCellbufLine(buf *uv.ScreenBuffer, i int) {
	clearLine(buf.Lines[i])
	buf.Touched[i] = nil
}

// clearLine sets all cells in the line to EmptyCell in place.
func clearLine(l uv.Line) {
	for i := range l {
		l[i] = uv.EmptyCell
	}
}
