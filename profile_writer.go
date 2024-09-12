package tea

import (
	"bytes"
	"fmt"
	"image/color"
	"io"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

// NewProfileWriter creates a new color profile writer that downgrades color
// sequences based on the detected color profile.
//
// If environ is nil, it will use os.Environ() to get the environment variables.
//
// It queries the given writer to determine if it supports ANSI escape codes.
// If it does, along with the given environment variables, it will determine
// the appropriate color profile to use for color formatting.
//
// This respects the NO_COLOR, CLICOLOR, and CLICOLOR_FORCE environment variables.
func NewProfileWriter(w io.Writer, environ []string) *ProfileWriter {
	return &ProfileWriter{
		Forward: w,
		Profile: detectColorProfile(w, environ),
	}
}

// When a color profile writer is created, it queries the given writer to
// determine if it supports ANSI escape codes. If it does, the color profile
// writer will write text with ANSI escape codes. If it does not, the profile
// writer will write text without ANSI escape codes.
//
// It also determines the appropriate color profile to use based on the
// capabilities of the underlying writer and environment.

// ProfileWriter represents a color profile writer that writes ANSI sequences
// to the underlying writer.
type ProfileWriter struct {
	Forward io.Writer
	Profile Profile
}

// Print writes the given text to the underlying writer.
func (w *ProfileWriter) Print(a ...any) (n int, err error) {
	return fmt.Fprint(w, a...)
}

// Println writes the given text to the underlying writer followed by a newline.
func (w *ProfileWriter) Println(a ...any) (n int, err error) {
	return fmt.Fprintln(w, a...)
}

// Printf writes the given text to the underlying writer with the given format.
func (w *ProfileWriter) Printf(format string, a ...any) (int, error) {
	return fmt.Fprintf(w, format, a...)
}

// Write writes the given text to the underlying writer.
func (w *ProfileWriter) Write(p []byte) (int, error) {
	switch w.Profile {
	case TrueColor:
		return w.Forward.Write(p)
	case NoTTY:
		return io.WriteString(w.Forward, ansi.Strip(string(p)))
	}

	convertColorAppend := func(c ansi.Color, sel colorSelector, pen *ansi.CsiSequence) {
		if c := w.Profile.Convert(c); c != nil {
			pen.Params = append(pen.Params, ansiColorToParams(c, sel)...)
		}
	}

	var buf bytes.Buffer
	var state byte
	pen := ansi.CsiSequence{Cmd: 'm'}

	parser := cellbuf.GetParser()
	defer cellbuf.PutParser(parser)

	for len(p) > 0 {
		parser.Reset()
		seq, width, read, newState := ansi.DecodeSequence(p, state, parser)

		if width == 0 && ansi.HasCsiPrefix(seq) && parser.Cmd == 'm' {
			pen.Params = pen.Params[:0]

			for j := 0; j < parser.ParamsLen; j++ {
				param := ansi.Param(parser.Params[j]).Param()
				switch param {
				case 30, 31, 32, 33, 34, 35, 36, 37: // 8-bit foreground color
					if w.Profile > ANSI {
						convertColorAppend(ansi.BasicColor(param-30), foreground, &pen) //nolint:gosec
						continue
					}
				case 39: // default foreground color
					if w.Profile > ANSI {
						continue
					}
				case 40, 41, 42, 43, 44, 45, 46, 47: // 8-bit background color
					if w.Profile > ANSI {
						convertColorAppend(ansi.BasicColor(param-40), background, &pen) //nolint:gosec
						continue
					}
				case 49: // default background color
					if w.Profile > ANSI {
						continue
					}
				case 90, 91, 92, 93, 94, 95, 96, 97: // 8-bit bright foreground color
					if w.Profile > ANSI {
						convertColorAppend(ansi.BasicColor(param-90+8), foreground, &pen) //nolint:gosec
						continue
					}
				case 100, 101, 102, 103, 104, 105, 106, 107: // 8-bit bright background color
					if w.Profile > ANSI {
						convertColorAppend(ansi.BasicColor(param-100+8), background, &pen) //nolint:gosec
						continue
					}
				case 59: // default underline color
					if w.Profile > ANSI {
						continue
					}
				case 38: // 16 or 24-bit foreground color
					fallthrough
				case 48: // 16 or 24-bit background color
					fallthrough
				case 58: // 16 or 24-bit underline color
					var sel colorSelector
					switch param {
					case 38:
						sel = foreground
					case 48:
						sel = background
					case 58:
						sel = underline
					}
					if c := readColor(&j, parser.Params); c != nil {
						switch c.(type) {
						case ansi.ExtendedColor:
							if w.Profile > ANSI256 {
								convertColorAppend(c, sel, &pen)
								continue
							}
						default:
							if w.Profile > TrueColor {
								convertColorAppend(c, sel, &pen)
								continue
							}
						}
						pen.Params = append(pen.Params, ansiColorToParams(c, sel)...)
						continue
					}
				default:
					pen.Params = append(pen.Params, param)
				}
			}

			if _, err := buf.Write(pen.Bytes()); err != nil {
				return 0, err
			}
		} else {
			if _, err := buf.Write(seq); err != nil {
				return 0, err
			}
		}

		p = p[read:]
		state = newState
	}

	return w.Forward.Write(buf.Bytes())
}

// WriteString writes the given text to the underlying writer.
func (w *ProfileWriter) WriteString(s string) (n int, err error) {
	return w.Write([]byte(s))
}

func readColor(idxp *int, params []int) (c ansi.Color) {
	i := *idxp
	paramsLen := len(params)
	if i > paramsLen-1 {
		return
	}
	// Note: we accept both main and subparams here
	switch param := ansi.Param(params[i+1]); param {
	case 2: // RGB
		if i > paramsLen-4 {
			return
		}
		c = color.RGBA{
			R: uint8(ansi.Param(params[i+2])), //nolint:gosec
			G: uint8(ansi.Param(params[i+3])), //nolint:gosec
			B: uint8(ansi.Param(params[i+4])), //nolint:gosec
			A: 0xff,
		}
		*idxp += 4
	case 5: // 256 colors
		if i > paramsLen-2 {
			return
		}
		c = ansi.ExtendedColor(ansi.Param(params[i+2])) //nolint:gosec
		*idxp += 2
	}
	return
}

type colorSelector uint8

const (
	foreground colorSelector = iota
	background
	underline
)

func ansiColorToParams(c ansi.Color, sel colorSelector) []int {
	switch c := c.(type) {
	case ansi.BasicColor:
		offset := 30
		if c >= ansi.BrightBlack {
			offset = 90
			c -= ansi.BrightBlack
		}
		switch sel {
		case foreground:
			return []int{offset + int(c)}
		case background:
			return []int{offset + 10 + int(c)}
		case underline:
			// NOTE: ANSI doesn't have underline colors, use ANSI256.
			return []int{58, 5, int(c)}
		}
	case ansi.ExtendedColor:
		switch sel {
		case foreground:
			return []int{38, 5, int(c)}
		case background:
			return []int{48, 5, int(c)}
		case underline:
			return []int{58, 5, int(c)}
		}
	default:
		r, g, b, _ := c.RGBA()
		r = r >> 8
		g = g >> 8
		b = b >> 8
		switch sel {
		case foreground:
			return []int{38, 2, int(r), int(g), int(b)}
		case background:
			return []int{48, 2, int(r), int(g), int(b)}
		case underline:
			return []int{58, 2, int(r), int(g), int(b)}
		}
	}
	return nil
}
