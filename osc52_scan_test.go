package tea

import (
	"bytes"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestSupportsOSC52(t *testing.T) {
	tests := []struct {
		name    string
		environ uv.Environ
		want    bool
	}{
		{"default", uv.Environ{"TERM=xterm-256color"}, true},
		{"apple terminal", uv.Environ{"TERM_PROGRAM=Apple_Terminal"}, false},
		{"apple terminal case", uv.Environ{"TERM_PROGRAM=APPLE_TERMINAL"}, false},
		{"iterm", uv.Environ{"TERM_PROGRAM=iTerm.app"}, true},
		{"wezterm", uv.Environ{"TERM_PROGRAM=WezTerm"}, true},
		{"tmux", uv.Environ{"TERM_PROGRAM=tmux"}, true},
		{"windows terminal", uv.Environ{"WT_SESSION=abc", "TERM=xterm-256color"}, true},
		{"ssh tty", uv.Environ{"TERM_PROGRAM=Apple_Terminal", "SSH_TTY=/dev/ttys001"}, true},
		{"ssh connection", uv.Environ{"TERM_PROGRAM=Apple_Terminal", "SSH_CONNECTION=1.2.3.4 5 6.7.8.9 7"}, true},
		{"linux console", uv.Environ{"TERM=linux"}, false},
		{"dumb terminal", uv.Environ{"TERM=dumb"}, false},
		{"old vte", uv.Environ{"VTE_VERSION=7000"}, false},
		{"new vte", uv.Environ{"VTE_VERSION=7600"}, true},
		{"invalid vte", uv.Environ{"VTE_VERSION=unknown"}, false},
		{"old konsole", uv.Environ{"KONSOLE_VERSION=220000"}, false},
		{"new konsole", uv.Environ{"KONSOLE_VERSION=230400"}, true},
		{"invalid konsole", uv.Environ{"KONSOLE_VERSION=x"}, false},
		{"ssh overrides vte", uv.Environ{"VTE_VERSION=7000", "SSH_TTY=/dev/pts/0"}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := supportsOSC52(test.environ); got != test.want {
				t.Errorf("supportsOSC52(%v) = %v, want %v", test.environ, got, test.want)
			}
		})
	}
}

func TestOSC52BridgeSet(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantOutput    string
		wantSelection byte
		wantContent   string
	}{
		{
			name:          "bel terminated",
			input:         "\x1b]52;c;aGVsbG8=\x07",
			wantOutput:    "",
			wantSelection: uv.SystemClipboard,
			wantContent:   "hello",
		},
		{
			name:          "st terminated",
			input:         "\x1b]52;c;aGVsbG8=\x1b\\",
			wantOutput:    "",
			wantSelection: uv.SystemClipboard,
			wantContent:   "hello",
		},
		{
			name:          "8-bit introducer and terminator",
			input:         "\x9d52;c;aGVsbG8=\x9c",
			wantOutput:    "",
			wantSelection: uv.SystemClipboard,
			wantContent:   "hello",
		},
		{
			name:          "surrounding output",
			input:         "before\x1b]52;c;aGVsbG8=\x07after",
			wantOutput:    "beforeafter",
			wantSelection: uv.SystemClipboard,
			wantContent:   "hello",
		},
		{
			name:          "primary selection",
			input:         "\x1b]52;p;aGVsbG8=\x07",
			wantOutput:    "",
			wantSelection: uv.PrimaryClipboard,
			wantContent:   "hello",
		},
		{
			name:          "empty payload",
			input:         "\x1b]52;c;\x07",
			wantOutput:    "",
			wantSelection: uv.SystemClipboard,
			wantContent:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			backend := &stubClipboardBackend{}
			output, _ := runOSC52Bridge(t, backend, test.input)

			if output != test.wantOutput {
				t.Errorf("output = %q, want %q", output, test.wantOutput)
			}
			calls := backend.setCallsFor()
			if len(calls) != 1 {
				t.Fatalf("backend set calls = %d, want 1", len(calls))
			}
			if calls[0].selection != test.wantSelection {
				t.Errorf("selection = %q, want %q", calls[0].selection, test.wantSelection)
			}
			if calls[0].content != test.wantContent {
				t.Errorf("content = %q, want %q", calls[0].content, test.wantContent)
			}
		})
	}
}

func TestOSC52BridgePassThrough(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"plain text", "hello, world"},
		{"osc title", "\x1b]0;a title\x07"},
		{"osc hyperlink", "\x1b]8;;https://example.com\x07link\x1b]8;;\x07"},
		{"malformed payload", "\x1b]52;c;!!!not-base64!!!\x07"},
		{"missing selection", "\x1b]52;;aGVsbG8=\x07"},
		{"cancelled sequence", "\x1b]52;c;aGVsbG8=\x18"},
		{"substituted sequence", "\x1b]52;c;aGVsbG8=\x1a"},
		{"other command", "\x1b]10;rgb:ffff/ffff/ffff\x07"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			backend := &stubClipboardBackend{}
			output, _ := runOSC52Bridge(t, backend, test.input)

			if output != test.input {
				t.Errorf("output = %q, want %q", output, test.input)
			}
			if calls := backend.setCallsFor(); len(calls) != 0 {
				t.Errorf("backend set calls = %d, want 0", len(calls))
			}
		})
	}
}

func TestOSC52BridgeRead(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "bel terminated",
			input: "before\x1b]52;c;?\x07after",
			want:  "beforeafter",
		},
		{
			name:  "st terminated",
			input: "\x1b]52;c;?\x1b\\",
			want:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			backend := &stubClipboardBackend{
				getValues: map[byte]string{uv.SystemClipboard: "hello"},
			}
			output, reply := runOSC52Bridge(t, backend, test.input)

			if output != test.want {
				t.Errorf("output = %q, want %q", output, test.want)
			}
			wantReply := "\x1b]52;c;aGVsbG8=\x07"
			if strings.Contains(test.input, "\x1b\\") {
				wantReply = "\x1b]52;c;aGVsbG8=\x1b\\"
			}
			if reply != wantReply {
				t.Errorf("reply = %q, want %q", reply, wantReply)
			}
			if got := backend.getCallsFor(); len(got) != 1 || got[0] != uv.SystemClipboard {
				t.Errorf("backend get calls = %v, want [%q]", got, uv.SystemClipboard)
			}
		})
	}
}

func TestOSC52BridgeSplit(t *testing.T) {
	seq := "\x1b]52;c;aGVsbG8=\x07"

	for i := 1; i < len(seq); i++ {
		backend := &stubClipboardBackend{}
		output, _ := runOSC52Bridge(t, backend, seq[:i], seq[i:])

		if output != "" {
			t.Errorf("split at %d: output = %q, want empty", i, output)
		}
		if calls := backend.setCallsFor(); len(calls) != 1 || calls[0].content != "hello" {
			t.Errorf("split at %d: set calls = %v, want one call with %q", i, calls, "hello")
		}
	}
}

func TestOSC52BridgeByteByByte(t *testing.T) {
	seq := "prefix\x1b]52;c;aGVsbG8=\x07suffix"
	chunks := make([]string, 0, len(seq))
	for _, r := range seq {
		chunks = append(chunks, string(r))
	}

	backend := &stubClipboardBackend{}
	output, _ := runOSC52Bridge(t, backend, chunks...)

	if output != "prefixsuffix" {
		t.Errorf("output = %q, want %q", output, "prefixsuffix")
	}
	if calls := backend.setCallsFor(); len(calls) != 1 || calls[0].content != "hello" {
		t.Errorf("set calls = %v, want one call with %q", calls, "hello")
	}
}

func TestOSC52BridgeHoldsPartialPrefix(t *testing.T) {
	var output, reply bytes.Buffer
	bridge := newOSC52Bridge(&stubClipboardBackend{}, &output, &reply)

	if _, err := bridge.Write([]byte("text\x1b]")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := output.String(); got != "text" {
		t.Errorf("output = %q, want %q (partial introducer should be held back)", got, "text")
	}

	backend := &stubClipboardBackend{}
	bridge.backend = backend
	if _, err := bridge.Write([]byte("52;c;aGVsbG8=\x07")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := output.String(); got != "text" {
		t.Errorf("output = %q, want %q", got, "text")
	}
	if calls := backend.setCallsFor(); len(calls) != 1 || calls[0].content != "hello" {
		t.Errorf("set calls = %v, want one call with %q", calls, "hello")
	}
}

func TestOSC52BridgeOversizedSequence(t *testing.T) {
	backend := &stubClipboardBackend{}
	huge := "\x1b]52;c;" + strings.Repeat("A", osc52MaxPending+1)

	output, _ := runOSC52Bridge(t, backend, huge)

	if !strings.HasPrefix(output, "\x1b]52;c;") {
		t.Errorf("oversized sequence should be forwarded untouched, got %q", output[:min(len(output), 32)])
	}
	if len(output) != len(huge) {
		t.Errorf("output length = %d, want %d", len(output), len(huge))
	}
	if calls := backend.setCallsFor(); len(calls) != 0 {
		t.Errorf("backend set calls = %d, want 0", len(calls))
	}
}

func TestOSC52BridgeFlush(t *testing.T) {
	t.Parallel()

	var output, reply bytes.Buffer
	bridge := newOSC52Bridge(&stubClipboardBackend{}, &output, &reply)

	// A trailing ESC that never became a sequence is held back...
	if _, err := bridge.Write([]byte("text\x1b")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := output.String(); got != "text" {
		t.Errorf("output before flush = %q, want %q", got, "text")
	}

	// ...and is written once the stream has ended.
	if err := bridge.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if got := output.String(); got != "text\x1b" {
		t.Errorf("output after flush = %q, want %q", got, "text\x1b")
	}
}

func TestOSC52BridgeBackendErrorsFailOpen(t *testing.T) {
	t.Parallel()

	t.Run("set error", func(t *testing.T) {
		t.Parallel()
		backend := &stubClipboardBackend{setErr: ErrClipboardUnavailable}
		input := "\x1b]52;c;aGVsbG8=\x07"
		output, _ := runOSC52Bridge(t, backend, input)
		if output != input {
			t.Errorf("output = %q, want %q", output, input)
		}
	})

	t.Run("get error", func(t *testing.T) {
		t.Parallel()
		backend := &stubClipboardBackend{getErr: ErrClipboardUnavailable}
		input := "\x1b]52;c;?\x07"
		output, _ := runOSC52Bridge(t, backend, input)
		if output != input {
			t.Errorf("output = %q, want %q", output, input)
		}
	})
}

// runOSC52Bridge runs the bridge over the given chunks and returns everything
// written to the output and reply writers.
func runOSC52Bridge(t *testing.T, backend ClipboardBackend, chunks ...string) (string, string) {
	t.Helper()

	var output, reply bytes.Buffer
	bridge := newOSC52Bridge(backend, &output, &reply)
	for _, chunk := range chunks {
		if _, err := bridge.Write([]byte(chunk)); err != nil {
			t.Fatalf("write %q: %v", chunk, err)
		}
	}
	return output.String(), reply.String()
}
