package tea

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

func (p *Program) consumeCursorStyleReport(event Msg) bool {
	report, ok := event.(uv.UnknownDcsEvent)
	if !ok {
		return false
	}
	style, valid := parseCursorStyleReport(string(report))
	if !valid {
		// A negative response ends capture, but remains available to the model.
		if isUnsupportedCursorStyleReport(string(report)) {
			if r, ok := p.renderer.(*cursedRenderer); ok {
				r.finishCursorStyleCapture()
			}
		}
		return false
	}
	r, ok := p.renderer.(*cursedRenderer)
	return ok && r.saveCursorStyle(style)
}

func isUnsupportedCursorStyleReport(report string) bool {
	for _, prefix := range []string{"\x1bP", "\x90"} {
		for _, suffix := range []string{"\x1b\\", "\x9c"} {
			if report == prefix+"0$r q"+suffix {
				return true
			}
		}
	}
	return false
}

func (s *cursedRenderer) finishCursorStyleCapture() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cursorStyleRequested {
		s.cursorStyleCaptureFinished = true
	}
}

// DECRQSS requests DECSCUSR before the renderer changes the cursor. Replies
// describe the style at query time, even if input processing happens later.
// Capture is best effort: unsupported terminals and replies not read before
// terminal release use the default fallback. DECRQSS has no request IDs, so
// application queries for the same status must not overlap this capture.
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

func (s *cursedRenderer) requestCursorStyleCapture() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.noInput || s.cursorStyleRequested {
		return false
	}
	s.cursorStyleRequested = true
	return true
}

func (s *cursedRenderer) saveCursorStyle(style int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cursorStyleRequested && !s.cursorStyleCaptureFinished {
		s.originalCursorStyle = style
		s.cursorStyleKnown = true
		s.cursorStyleCaptureFinished = true
		return true
	}
	return false
}
