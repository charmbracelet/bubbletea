package tea

import (
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/xo/terminfo"
)

// detectColorProfile returns the color profile based on the terminal output,
// and environment variables. This respects NO_COLOR, CLICOLOR, and
// CLICOLOR_FORCE environment variables.
//
// The rules as follows:
//   - TERM=dumb is always treated as NoTTY unless CLICOLOR_FORCE=1 is set.
//   - If COLORTERM=truecolor, and the profile is not NoTTY, it gest upgraded to TrueColor.
//   - Using any 256 color terminal (e.g. TERM=xterm-256color) will set the profile to ANSI256.
//   - Using any color terminal (e.g. TERM=xterm-color) will set the profile to ANSI.
//   - Using CLICOLOR=1 without TERM defined should be treated as ANSI if the
//     output is a terminal.
//   - NO_COLOR takes precedence over CLICOLOR/CLICOLOR_FORCE, and will disable
//     colors but not text decoration, i.e. bold, italic, faint, etc.
//
// See https://no-color.org/ and https://bixense.com/clicolors/ for more information.
func detectColorProfile(output io.Writer, environ []string) (p Profile) {
	out, ok := output.(term.File)
	isatty := ok && term.IsTerminal(out.Fd())
	return colorProfile(isatty, environ)
}

func colorProfile(isatty bool, environ []string) (p Profile) {
	env := environMap(environ)
	envProfile := envColorProfile(env)

	// Start with the environment profile.
	p = envProfile

	term := strings.ToLower(env["TERM"])
	isDumb := term == "dumb"

	// Check if the output is a terminal.
	// Treat dumb terminals as NoTTY
	if !isatty || isDumb {
		p = NoTTY
	}

	if envNoColor(env) {
		if p < Ascii {
			p = Ascii
		}
		return
	}

	if cliColorForced(env) {
		if p > ANSI {
			p = ANSI
		}
		if envProfile < p {
			p = envProfile
		}

		return
	}

	if cliColor(env) {
		if isatty && !isDumb && p > ANSI {
			p = ANSI
		}
	}

	return p
}

// envNoColor returns true if the environment variables explicitly disable color output
// by setting NO_COLOR (https://no-color.org/).
func envNoColor(env map[string]string) bool {
	noColor, _ := strconv.ParseBool(env["NO_COLOR"])
	return noColor
}

func cliColor(env map[string]string) bool {
	cliColor, _ := strconv.ParseBool(env["CLICOLOR"])
	return cliColor
}

func cliColorForced(env map[string]string) bool {
	cliColorForce, _ := strconv.ParseBool(env["CLICOLOR_FORCE"])
	return cliColorForce
}

func colorTerm(env map[string]string) bool {
	colorTerm := strings.ToLower(env["COLORTERM"])
	return colorTerm == "truecolor" || colorTerm == "24bit" ||
		colorTerm == "yes" || colorTerm == "true"
}

// envColorProfile returns infers the color profile from the environment.
func envColorProfile(env map[string]string) (p Profile) {
	p = Ascii // Default to Ascii
	if isCloudShell, _ := strconv.ParseBool(env["GOOGLE_CLOUD_SHELL"]); isCloudShell {
		p = TrueColor
		return
	}

	term := strings.ToLower(env["TERM"])
	switch term {
	case "", "dumb":
		p = NoTTY
	}

	if colorTerm(env) {
		p = TrueColor
		return
	}

	switch term {
	case "alacritty", "contour", "wezterm", "xterm-ghostty", "xterm-kitty":
		p = TrueColor
		return
	case "linux":
		if p > ANSI {
			p = ANSI
		}
	}

	if strings.Contains(term, "256color") && p > ANSI256 {
		p = ANSI256
	}
	if strings.Contains(term, "color") && p > ANSI {
		p = ANSI
	}
	if strings.Contains(term, "ansi") && p > ANSI {
		p = ANSI
	}

	if ti, err := terminfo.Load(term); err == nil {
		extbools := ti.ExtBoolCapsShort()
		if _, ok := extbools["RGB"]; ok {
			p = TrueColor
			return
		}

		if _, ok := extbools["Tc"]; ok {
			p = TrueColor
			return
		}

		nums := ti.NumCapsShort()
		if colors, ok := nums["colors"]; ok {
			if colors >= 0x1000000 {
				p = TrueColor
				return
			} else if colors >= 0x100 && p > ANSI256 {
				p = ANSI256
			} else if colors >= 0x10 && p > ANSI {
				p = ANSI
			}
		}
	}

	return
}

// environMap converts an environment slice to a map.
func environMap(environ []string) map[string]string {
	m := make(map[string]string, len(environ))
	for _, e := range environ {
		parts := strings.SplitN(e, "=", 2)
		var value string
		if len(parts) == 2 {
			value = parts[1]
		}
		m[parts[0]] = value
	}
	return m
}
