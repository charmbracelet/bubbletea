package tea

import (
	"fmt"
	"io"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// generateStyledContent creates ANSI-styled content simulating a TUI viewport.
// Each line has foreground color, bold text, and a reset — similar to real TUI output.
func generateStyledContent(width, height, offset int) string {
	var sb strings.Builder
	for y := 0; y < height; y++ {
		lineNum := offset + y
		prefix := fmt.Sprintf("\x1b[38;2;200;200;200m\x1b[48;2;30;30;46m%4d │ ", lineNum)
		body := fmt.Sprintf("Line content for row %d with some text that fills the width", lineNum)
		visibleLen := 7 + len(body)
		if visibleLen < width {
			body += strings.Repeat(" ", width-visibleLen)
		} else if visibleLen > width {
			body = body[:width-7]
		}
		sb.WriteString(prefix)
		sb.WriteString(body)
		sb.WriteString("\x1b[0m")
		if y < height-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

func BenchmarkFlushScroll(b *testing.B) {
	const width, height = 200, 50
	const scrollStep = 3

	env := []string{"TERM=xterm-256color", "COLORTERM=truecolor"}
	r := newCursedRenderer(io.Discard, env, width, height)
	r.syncdUpdates = false

	view := View{
		Content:   generateStyledContent(width, height, 0),
		AltScreen: true,
	}
	r.render(view)
	_ = r.flush(false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		offset := (i + 1) * scrollStep
		view.Content = generateStyledContent(width, height, offset)
		r.render(view)
		_ = r.flush(false)
	}
}

func BenchmarkFlushStatic(b *testing.B) {
	const width, height = 200, 50

	env := []string{"TERM=xterm-256color", "COLORTERM=truecolor"}
	r := newCursedRenderer(io.Discard, env, width, height)
	r.syncdUpdates = false

	view := View{
		Content:   generateStyledContent(width, height, 0),
		AltScreen: true,
	}
	r.render(view)
	_ = r.flush(false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.render(view)
		_ = r.flush(false)
	}
}

func BenchmarkFlushFullChange(b *testing.B) {
	const width, height = 200, 50

	env := []string{"TERM=xterm-256color", "COLORTERM=truecolor"}
	r := newCursedRenderer(io.Discard, env, width, height)
	r.syncdUpdates = false

	view := View{
		Content:   generateStyledContent(width, height, 0),
		AltScreen: true,
	}
	r.render(view)
	_ = r.flush(false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		view.Content = generateStyledContent(width, height, (i+1)*height)
		r.render(view)
		_ = r.flush(false)
	}
}

// generatePartialScrollContent creates content simulating steiner's layout:
// topChromeLines constant lines at top, contentLines scrolling, bottomChromeLines
// constant at bottom. This mirrors real TUI apps that have padding/header at the
// top and status/input at the bottom.
func generatePartialScrollContent(width, topChromeLines, contentLines, bottomChromeLines, offset int) string {
	var sb strings.Builder
	height := topChromeLines + contentLines + bottomChromeLines
	for y := 0; y < height; y++ {
		var line string
		if y < topChromeLines {
			line = fmt.Sprintf("\x1b[48;2;30;30;46m%s\x1b[0m", strings.Repeat(" ", width))
		} else if y < topChromeLines+contentLines {
			lineNum := offset + y - topChromeLines
			prefix := fmt.Sprintf("\x1b[38;2;200;200;200m\x1b[48;2;30;30;46m%4d │ ", lineNum)
			body := fmt.Sprintf("Line content for row %d with some text", lineNum)
			visibleLen := 7 + len(body)
			if visibleLen < width {
				body += strings.Repeat(" ", width-visibleLen)
			}
			line = prefix + body + "\x1b[0m"
		} else {
			switch y - topChromeLines - contentLines {
			case 0:
				line = strings.Repeat("─", width)
			case 1:
				line = fmt.Sprintf("> %s", strings.Repeat(" ", width-2))
			default:
				line = fmt.Sprintf("\x1b[7m steiner \x1b[0m%s", strings.Repeat(" ", width-9))
			}
		}
		sb.WriteString(line)
		if y < height-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

func BenchmarkFlushScrollPartial(b *testing.B) {
	const width = 200
	const topChrome = 1
	const contentLines = 39
	const bottomChrome = 10
	const scrollStep = 3

	height := topChrome + contentLines + bottomChrome
	env := []string{"TERM=xterm-256color", "COLORTERM=truecolor"}
	r := newCursedRenderer(io.Discard, env, width, height)
	r.syncdUpdates = false

	view := View{
		Content:   generatePartialScrollContent(width, topChrome, contentLines, bottomChrome, 0),
		AltScreen: true,
	}
	r.render(view)
	_ = r.flush(false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		offset := (i + 1) * scrollStep
		view.Content = generatePartialScrollContent(width, topChrome, contentLines, bottomChrome, offset)
		r.render(view)
		_ = r.flush(false)
	}
}

// generateSuffixContent creates ANSI-styled content where each line ends with a
// scrollbar character that varies per frame (thumbPos controls which line gets
// the thumb). This simulates a viewport with a live scrollbar: the content
// prefix is stable across frames so prefix-match fires, but exact equality
// fails because the thumb position shifts.
func generateSuffixContent(width, height, offset int, thumbPos int) string {
	var sb strings.Builder
	for y := 0; y < height; y++ {
		lineNum := offset + y
		prefix := fmt.Sprintf("\x1b[38;2;200;200;200m\x1b[48;2;30;30;46m%4d │ ", lineNum)
		body := fmt.Sprintf("Line content for row %d with some text that fills the width", lineNum)
		visibleLen := 7 + len(body)
		if visibleLen < width-3 {
			body += strings.Repeat(" ", width-3-visibleLen)
		}
		var scrollChar string
		if y == thumbPos {
			scrollChar = "\x1b[7m▓\x1b[0m" // reverse-video thumb
		} else {
			scrollChar = "│"
		}
		sb.WriteString(prefix)
		sb.WriteString(body)
		sb.WriteString(scrollChar)
		sb.WriteString("\x1b[0m")
		if y < height-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

func BenchmarkFlushScrollWithSuffix(b *testing.B) {
	const width, height = 200, 50
	const scrollStep = 3

	env := []string{"TERM=xterm-256color", "COLORTERM=truecolor"}
	r := newCursedRenderer(io.Discard, env, width, height)
	r.syncdUpdates = false

	view := View{
		Content:   generateSuffixContent(width, height, 0, 25),
		AltScreen: true,
	}
	r.render(view)
	_ = r.flush(false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		offset := (i + 1) * scrollStep
		thumbPos := (offset * height / (height * 10)) % height // moves slowly
		view.Content = generateSuffixContent(width, height, offset, thumbPos)
		r.render(view)
		_ = r.flush(false)
	}
}

func BenchmarkDrawOnly(b *testing.B) {
	const width, height = 200, 50

	content := generateStyledContent(width, height, 0)
	cellbuf := uv.NewScreenBuffer(width, height)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cellbuf.Clear()
		s := uv.NewStyledString(content)
		s.Draw(cellbuf, cellbuf.Bounds())
	}
}
