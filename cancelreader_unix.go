//go:build solaris
// +build solaris

// nolint:revive
package tea

import (
	"io"
)

// newCancelReader returns a reader and a cancel function. If the input reader
// is an *os.File, the cancel function can be used to interrupt a blocking call
// read call. In this case, the cancel function returns true if the call was
// cancelled successfully. If the input reader is not a *os.File or the file
// descriptor is 1024 or larger, the cancel function does nothing and always
// returns false. The generic unix implementation is based on the posix select
// syscall.
func newCancelReader(reader io.Reader) (cancelReader, error) {
	return newSelectCancelReader(reader)
}
