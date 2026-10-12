package tea

import "strings"

// DECRQSS requests DECSCUSR before the renderer changes the cursor. Replies
// describe the style at query time, even if input processing happens later.
const requestCursorStyle = "\x1bP$q q\x1b\\"

func parseCursorStyleReport(report string) (int, bool) {
	switch {
	case strings.HasPrefix(report, "\x1bP"):
		report = report[2:]
	case strings.HasPrefix(report, "\x90"):
		report = report[1:]
	default:
		return 0, false
	}
	switch {
	case strings.HasSuffix(report, "\x1b\\"):
		report = report[:len(report)-2]
	case strings.HasSuffix(report, "\x9c"):
		report = report[:len(report)-1]
	default:
		return 0, false
	}
	if len(report) != 6 || report[:3] != "1$r" || report[4:] != " q" {
		return 0, false
	}
	style := int(report[3] - '0')
	return style, style >= 0 && style <= 6
}

func (s *cursedRenderer) saveCursorStyle(style int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.cursorStyleKnown {
		s.originalCursorStyle = style
		s.cursorStyleKnown = true
	}
}
