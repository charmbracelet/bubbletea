//go:build !windows
// +build !windows

package tea

import (
	"context"
	"io"
)

func readInputs(ctx context.Context, msgs chan<- Msg, input io.Reader) error {
	return readAnsiInputs(ctx, msgs, input)
}
