//go:build !windows
// +build !windows

package tea

import (
	"fmt"
	"io"

	"github.com/muesli/cancelreader"
)

func newInputReader(r io.Reader, _ bool) (cancelreader.CancelReader, error) {
	cr, err := cancelreader.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("bubbletea: error creating cancel reader: %w", err)
	}
	return cr, nil
}
