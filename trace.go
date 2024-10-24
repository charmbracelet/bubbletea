package tea

import "os"

// tracerMask is a bitmask that represents the different types of tracing that
// can be enabled.
type tracerMask int

// Constants that represent the different types of tracing that can be enabled.
const (
	traceOutput tracerMask = 1 << iota
	traceInput
)

// tracer is a type that can be used to trace ANSI sequences and I/O of a
// program.
type tracer struct {
	file *os.File
	op   string
	mask tracerMask
}

// newTracer returns a new tracer for the given path and mask options.
func newTracer(fp string, masks ...tracerMask) (tracer, error) {
	var mask tracerMask
	for _, m := range masks {
		mask |= m
	}

	t := tracer{op: fp, mask: mask}
	f, err := LogToFile(fp, "bubbletea")
	if err != nil {
		return t, err
	}

	t.file = f
	return t, nil
}

// withOutput returns a new tracer that traces output.
func (t tracer) withOutput() tracer {
	t.mask |= traceOutput
	return t
}

// withInput returns a new tracer that traces input.
func (t tracer) withInput() tracer {
	t.mask |= traceInput
	return t
}
