package tea

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestLogToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log.txt")
	prefix := "logprefix"
	f, err := LogToFile(path, prefix)
	if err != nil {
		t.Error(err)
	}
	log.SetFlags(log.Lmsgprefix)
	log.Println("some test log")
	if closeErr := f.Close(); closeErr != nil {
		t.Error(closeErr)
	}
	out, err := os.ReadFile(path)
	if err != nil {
		t.Error(err)
	}
	if string(out) != prefix+" some test log\n" {
		t.Fatalf("wrong log msg: %q", string(out))
	}
}

func TestLogToFileWithPrefixSpacing(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{"empty", "", ""},
		{"missing space", "prefix", "prefix "},
		{"ASCII space", "prefix ", "prefix "},
		{"non-breaking space", "prefix\u00a0", "prefix\u00a0"},
		{"ideographic space", "prefix\u3000", "prefix\u3000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(io.Discard, "", 0)
			f, err := LogToFileWith(filepath.Join(t.TempDir(), "log.txt"), tt.prefix, logger)
			if err != nil {
				t.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}

			if got := logger.Prefix(); got != tt.want {
				t.Fatalf("prefix = %q; want %q", got, tt.want)
			}
		})
	}
}
