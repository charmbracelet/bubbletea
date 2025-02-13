package tea

import (
	"io"
	"log"
	"sync"
)

// safeWriter is a thread-safe writer.
type safeWriter struct {
	w     io.Writer
	mu    sync.Mutex
	trace bool
}

var _ io.Writer = &safeWriter{}

// newSafeWriter returns a new safeWriter.
func newSafeWriter(w io.Writer) *safeWriter {
	return &safeWriter{w: w}
}

// Writer returns the underlying writer.
func (w *safeWriter) Writer() io.Writer {
	return w.w
}

// Write writes to the underlying writer.
func (w *safeWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.trace {
		log.Printf("output %q", p)
	}
	return w.w.Write(p)
}
