package tea

import (
	"io"
	"log"
)

// traceWriter is a writer that logs writes to an underlying writer.
type traceWriter struct {
	io.Writer
	trace bool
}

// Write writes to the underlying writer and logs the write if tracing is enabled.
func (w *traceWriter) Write(p []byte) (n int, err error) {
	if w.trace {
		log.Printf("output: %q", p)
	}
	return w.Writer.Write(p)
}
