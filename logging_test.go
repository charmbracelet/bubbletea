package tea

import (
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
	if err := f.Close(); err != nil {
		t.Error(err)
	}
	out, err := os.ReadFile(path)
	if err != nil {
		t.Error(err)
	}
	if string(out) != prefix+" some test log\n" {
		t.Fatalf("wrong log msg: %q", string(out))
	}
}
