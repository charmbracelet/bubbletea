package tea

import (
	"fmt"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// makeLines returns a slice of n strings starting from offset, e.g.
// makeLines(0, 4) → ["line0", "line1", "line2", "line3"].
func makeLines(offset, n int) []string {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("line%d", offset+i)
	}
	return lines
}

// makeSuffixLines returns n lines like "content{offset+i}" padded to ~100 bytes
// plus the given suffix. Used to simulate scrollbar decoration that changes
// between frames while content prefix is stable.
func makeSuffixLines(offset, n int, suffix string) []string {
	lines := make([]string, n)
	for i := range lines {
		base := fmt.Sprintf("content%d", offset+i)
		// Pad base to ~100 bytes so prefix comparison has room.
		if len(base) < 100 {
			base += strings.Repeat(" ", 100-len(base))
		}
		lines[i] = base + suffix
	}
	return lines
}

func TestLinesMatchForShift(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{
			name: "exact match short strings",
			a:    "hello",
			b:    "hello",
			want: true, // fast path: a==b
		},
		{
			name: "both empty",
			a:    "",
			b:    "",
			want: true, // fast path: a==b
		},
		{
			name: "exact match longer strings",
			a:    strings.Repeat("a", 40),
			b:    strings.Repeat("a", 40),
			want: true, // fast path: a==b
		},
		{
			name: "too short to prefix-compare different strings",
			a:    strings.Repeat("x", 30),
			b:    strings.Repeat("y", 30),
			want: false, // n=30 <= maxTrailingDiff=40; fast path fails; returns false
		},
		{
			name: "prefix differs content completely different",
			a:    strings.Repeat("a", 50) + "TAIL",
			b:    strings.Repeat("b", 50) + "TAIL",
			want: false, // prefix a[:14]="aa..." != b[:14]="bb..."
		},
		{
			name: "suffix-only difference within trailing margin",
			a:    strings.Repeat("a", 100) + "SCROLLBAR1",
			b:    strings.Repeat("a", 100) + "SCROLLBAR2",
			want: true, // n=110, compareLen=70; a[:70]==b[:70]; suffixes differ but within 40 bytes
		},
		{
			name: "prefix match suffix change exactly at boundary",
			a:    strings.Repeat("z", 80) + strings.Repeat("X", 40),
			b:    strings.Repeat("z", 80) + strings.Repeat("Y", 40),
			want: true, // n=120, compareLen=80; a[:80]==b[:80]
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := linesMatchForShift(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("linesMatchForShift(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDetectContentShift(t *testing.T) {
	tests := []struct {
		name            string
		old             []string
		new             []string
		wantShift       int
		wantRegionStart int
		wantMatchCount  int
	}{
		{
			name:           "no change",
			old:            makeLines(0, 20),
			new:            makeLines(0, 20),
			wantShift:      0,
			wantMatchCount: 0,
		},
		{
			name:           "scroll up by 1",
			old:            makeLines(0, 20),
			new:            append(makeLines(1, 19), "newline"),
			wantShift:      1,
			wantMatchCount: 19,
		},
		{
			name:           "scroll up by 2",
			old:            makeLines(0, 20),
			new:            append(makeLines(2, 18), "new0", "new1"),
			wantShift:      2,
			wantMatchCount: 18,
		},
		{
			name:           "scroll down by 1",
			old:            makeLines(1, 20),
			new:            append([]string{"line0"}, makeLines(1, 19)...),
			wantShift:      -1,
			wantMatchCount: 19,
		},
		{
			name:           "completely different",
			old:            makeLines(0, 20),
			new:            makeLines(100, 20),
			wantShift:      0,
			wantMatchCount: 0,
		},
		{
			name:           "too short",
			old:            []string{"a", "b", "c"},
			new:            []string{"b", "c", "d"},
			wantShift:      0,
			wantMatchCount: 0,
		},
		{
			name:           "different lengths",
			old:            []string{"a", "b", "c", "d"},
			new:            []string{"a", "b", "c"},
			wantShift:      0,
			wantMatchCount: 0,
		},
		{
			name: "scroll up with fixed chrome at bottom",
			old: func() []string {
				s := make([]string, 20)
				for i := 0; i < 17; i++ {
					s[i] = fmt.Sprintf("content%d", i)
				}
				s[17], s[18], s[19] = "chrome-divider", "chrome-input", "chrome-status"
				return s
			}(),
			new: func() []string {
				s := make([]string, 20)
				for i := 0; i < 16; i++ {
					s[i] = fmt.Sprintf("content%d", i+1)
				}
				s[16] = "content-new"
				s[17], s[18], s[19] = "chrome-divider", "chrome-input", "chrome-status"
				return s
			}(),
			wantShift:      1,
			wantMatchCount: 16,
		},
		{
			name: "no shift when too few content lines match",
			old: func() []string {
				s := make([]string, 20)
				for i := range s {
					s[i] = fmt.Sprintf("content%d", i)
				}
				return s
			}(),
			new: func() []string {
				s := make([]string, 20)
				for i := 0; i < 4; i++ {
					s[i] = fmt.Sprintf("content%d", i+1)
				}
				for i := 4; i < 20; i++ {
					s[i] = fmt.Sprintf("different%d", i)
				}
				return s
			}(),
			wantShift:      0,
			wantMatchCount: 0,
		},
		{
			name:           "scroll up with scrollbar suffix",
			old:            makeSuffixLines(0, 20, "│▓│"),
			new:            append(makeSuffixLines(1, 19, "│░│"), makeSuffixLines(20, 1, "│░│")...),
			wantShift:      1,
			wantMatchCount: 19,
		},
		{
			name:           "scroll down with scrollbar suffix",
			old:            append(makeSuffixLines(1, 19, "│▓│"), makeSuffixLines(20, 1, "│▓│")...),
			new:            makeSuffixLines(0, 20, "│░│"),
			wantShift:      -1,
			wantMatchCount: 19,
		},
		{
			// Steiner pattern: 1 constant top chrome + 16 scrolling + 3 bottom chrome.
			// Chrome is stripped, shift detected in the scrolling region.
			name: "scroll up with chrome at top and bottom",
			old: func() []string {
				s := make([]string, 20)
				s[0] = "padded-empty-line"
				for i := 1; i < 17; i++ {
					s[i] = fmt.Sprintf("content%d", i-1)
				}
				s[17], s[18], s[19] = "chrome-divider", "chrome-input", "chrome-status"
				return s
			}(),
			new: func() []string {
				s := make([]string, 20)
				s[0] = "padded-empty-line"
				for i := 1; i < 16; i++ {
					s[i] = fmt.Sprintf("content%d", i) // shifted by 1
				}
				s[16] = "content-new"
				s[17], s[18], s[19] = "chrome-divider", "chrome-input", "chrome-status"
				return s
			}(),
			wantShift:       1,
			wantRegionStart: 1,
			wantMatchCount:  15,
		},
		{
			// Scroll down with chrome at both ends.
			name: "scroll down with chrome at top and bottom",
			old: func() []string {
				s := make([]string, 20)
				s[0] = "padded-empty-line"
				for i := 1; i < 17; i++ {
					s[i] = fmt.Sprintf("content%d", i)
				}
				s[17], s[18], s[19] = "chrome-divider", "chrome-input", "chrome-status"
				return s
			}(),
			new: func() []string {
				s := make([]string, 20)
				s[0] = "padded-empty-line"
				s[1] = "content-new-top"
				for i := 2; i < 17; i++ {
					s[i] = fmt.Sprintf("content%d", i-1)
				}
				s[17], s[18], s[19] = "chrome-divider", "chrome-input", "chrome-status"
				return s
			}(),
			wantShift:       -1,
			wantRegionStart: 1,
			wantMatchCount:  15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotShift, gotRegionStart, gotMatch := detectContentShift(tt.old, tt.new)
			if gotShift != tt.wantShift || gotRegionStart != tt.wantRegionStart || gotMatch != tt.wantMatchCount {
				t.Errorf("detectContentShift() = (%d, %d, %d), want (%d, %d, %d)",
					gotShift, gotRegionStart, gotMatch, tt.wantShift, tt.wantRegionStart, tt.wantMatchCount)
			}
		})
	}
}

func TestShiftCellbufRegion(t *testing.T) {
	const width = 5
	const height = 10

	// fillLine sets every cell in row to a string tag (single char repeated).
	fillLine := func(buf *uv.ScreenBuffer, row int, tag string) {
		for i := range buf.Lines[row] {
			buf.Lines[row][i] = uv.Cell{Content: tag, Width: 1}
		}
	}

	// firstCellContent returns the Content field of the first cell in a row.
	firstCellContent := func(buf *uv.ScreenBuffer, row int) string {
		return buf.Lines[row][0].Content
	}

	// isBlankLine reports whether all cells in row match EmptyCell.
	isBlankLine := func(buf *uv.ScreenBuffer, row int) bool {
		for _, cell := range buf.Lines[row] {
			if cell != uv.EmptyCell {
				return false
			}
		}
		return true
	}

	t.Run("shift up within region leaves chrome untouched", func(t *testing.T) {
		buf := uv.NewScreenBuffer(width, height)
		for i := 0; i < 7; i++ {
			fillLine(&buf, i, string(rune('A'+i)))
		}
		for i := 7; i < 10; i++ {
			fillLine(&buf, i, "X")
		}

		shiftCellbufRegion(&buf, 0, 7, 2)

		for i := 0; i < 5; i++ {
			want := string(rune('C' + i))
			if got := firstCellContent(&buf, i); got != want {
				t.Errorf("row %d: got %q, want %q", i, got, want)
			}
		}
		for i := 5; i < 7; i++ {
			if !isBlankLine(&buf, i) {
				t.Errorf("row %d: expected blank after shift up, got non-blank", i)
			}
		}
		for i := 7; i < 10; i++ {
			if got := firstCellContent(&buf, i); got != "X" {
				t.Errorf("chrome row %d: got %q, want \"X\"", i, got)
			}
		}
	})

	t.Run("shift down within region leaves chrome untouched", func(t *testing.T) {
		buf := uv.NewScreenBuffer(width, height)
		for i := 0; i < 7; i++ {
			fillLine(&buf, i, string(rune('A'+i)))
		}
		for i := 7; i < 10; i++ {
			fillLine(&buf, i, "X")
		}

		shiftCellbufRegion(&buf, 0, 7, -2)

		for i := 0; i < 2; i++ {
			if !isBlankLine(&buf, i) {
				t.Errorf("row %d: expected blank after shift down, got non-blank", i)
			}
		}
		for i := 2; i < 7; i++ {
			want := string(rune('A' + i - 2))
			if got := firstCellContent(&buf, i); got != want {
				t.Errorf("row %d: got %q, want %q", i, got, want)
			}
		}
		for i := 7; i < 10; i++ {
			if got := firstCellContent(&buf, i); got != "X" {
				t.Errorf("chrome row %d: got %q, want \"X\"", i, got)
			}
		}
	})

	t.Run("shift with regionStart skips top chrome", func(t *testing.T) {
		buf := uv.NewScreenBuffer(width, height)
		fillLine(&buf, 0, "T") // top chrome
		for i := 1; i < 8; i++ {
			fillLine(&buf, i, string(rune('A'+i-1)))
		}
		for i := 8; i < 10; i++ {
			fillLine(&buf, i, "X")
		}

		// Shift region [1,8) up by 2. Top chrome at 0 and bottom chrome at 8-9 untouched.
		shiftCellbufRegion(&buf, 1, 8, 2)

		if got := firstCellContent(&buf, 0); got != "T" {
			t.Errorf("top chrome row 0: got %q, want \"T\"", got)
		}
		for i := 1; i < 6; i++ {
			want := string(rune('A' + i + 1))
			if got := firstCellContent(&buf, i); got != want {
				t.Errorf("row %d: got %q, want %q", i, got, want)
			}
		}
		for i := 6; i < 8; i++ {
			if !isBlankLine(&buf, i) {
				t.Errorf("row %d: expected blank after shift up, got non-blank", i)
			}
		}
		for i := 8; i < 10; i++ {
			if got := firstCellContent(&buf, i); got != "X" {
				t.Errorf("bottom chrome row %d: got %q, want \"X\"", i, got)
			}
		}
	})

	t.Run("nil Touched for all lines after shift", func(t *testing.T) {
		buf := uv.NewScreenBuffer(width, height)
		for i := range buf.Touched {
			buf.Touched[i] = &uv.LineData{FirstCell: 0, LastCell: width - 1}
		}

		shiftCellbufRegion(&buf, 0, 7, 1)

		// After shift, ALL lines must have nil Touched so updateHashmap and
		// transformLine skip them entirely.
		for i := range buf.Touched {
			if buf.Touched[i] != nil {
				t.Errorf("Touched[%d] = %v, want nil", i, buf.Touched[i])
			}
		}
	})
}
