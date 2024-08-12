//go:build !windows
// +build !windows

package tea

import (
	"io"

	"github.com/muesli/cancelreader"
)

func newCancelreader(r io.Reader) (cancelreader.CancelReader, error) {
	return cancelreader.NewReader(r)
}
