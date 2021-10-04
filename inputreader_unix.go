//go:build solaris
// +build solaris

// nolint:revive
package tea

import (
	"io"
)

// newInputReader returns a cancelable input reader. If the passed reader is an
// *os.File, the cancel method can be used to interrupt a blocking call read
// call. In this case, the cancel method returns true if the call was cancelled
// successfully. If the input reader is not a *os.File or the file descriptor is
// 1024 or larger, the cancel method does nothing and always returns false. The
// generic Unix implementation is based on the POSIX select syscall.
func newInputReader(reader io.Reader) (inputReader, error) {
	return newSelectInputReader(reader)
}
