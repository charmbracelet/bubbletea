package tea

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/muesli/termenv"
)

const (
	defaultFramerate = time.Millisecond * 16
)

type renderer struct {
	out           io.Writer
	buf           bytes.Buffer
	framerate     time.Duration
	ticker        *time.Ticker
	mtx           sync.Mutex
	done          chan struct{}
	lastRender    string
	linesRendered int
}

func newRenderer(out io.Writer) *renderer {
	return &renderer{
		out:       out,
		framerate: defaultFramerate,
	}
}

func (r *renderer) start() {
	if r.ticker == nil {
		r.ticker = time.NewTicker(r.framerate)
	}
	r.done = make(chan struct{})
	go r.listen()
}

func (r *renderer) stop() {
	r.flush()
	r.done <- struct{}{}
}

func (r *renderer) listen() {
	for {
		select {
		case <-r.ticker.C:
			if r.ticker != nil {
				r.flush()
			}
		case <-r.done:
			r.mtx.Lock()
			r.ticker.Stop()
			r.ticker = nil
			r.mtx.Unlock()
			close(r.done)
			return
		}
	}
}

func (r *renderer) flush() {
	if r.buf.Len() == 0 || r.buf.String() == r.lastRender {
		// Nothing to do
		return
	}

	// We have an opportunity here to limit the rendering to the terminal width
	// and height, but this would mean a few things:
	//
	// 1) We'd need to maintain the terminal dimensions internally and listen
	// for window size changes.
	//
	// 2) We'd need to measure the width of lines, accounting for double-byte
	// widths. We'd use something like go-runewidth
	// (http://github.com/mattn/go-runewidth).
	//
	// 3) We'd need to measure the width of lines excluding ANSI escape
	// sequences and break lines in the right places accordingly.
	//
	// Because of the way this would complicate the renderer, this may not be
	// the place to do that.

	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.linesRendered > 0 {
		termenv.ClearLines(r.linesRendered)
	}
	r.linesRendered = 0

	var out bytes.Buffer
	for _, b := range r.buf.Bytes() {
		if b == '\n' {
			r.linesRendered++
			out.Write([]byte("\r\n"))
		} else {
			_, _ = out.Write([]byte{b})
		}
	}

	_, _ = r.out.Write(out.Bytes())
	r.lastRender = r.buf.String()
	r.buf.Reset()
}

func (w *renderer) write(s string) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	w.buf.Reset()
	w.buf.WriteString(s)
}
