package tea

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log.txt")
	prefix := "logprefix"
	f, err := LogToFile(path, prefix)
	assert.NoError(t, err)

	log.SetFlags(log.Lmsgprefix)
	log.Println("some test log")
	assert.NoError(t, f.Close())

	out, err := os.ReadFile(path)
	assert.NoError(t, err)

	assert.Equal(t, prefix+" some test log\n", string(out))
}
