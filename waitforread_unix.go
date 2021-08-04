// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package tea

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

var errAborted = fmt.Errorf("select aborted")

func waitForRead(reader *os.File, abort *os.File) error {
	readerFd := int(reader.Fd())
	abortFd := int(abort.Fd())

	maxFd := readerFd
	if abortFd > maxFd {
		maxFd = abortFd
	}

	if maxFd >= 1024 {
		return fmt.Errorf("cannot select on file descriptor %d which is larger than 1024", maxFd)
	}

	fdSet := &unix.FdSet{}
	fdSet.Set(int(reader.Fd()))
	fdSet.Set(int(abort.Fd()))

	_, err := unix.Select(maxFd+1, fdSet, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("select: %w", err)
	}

	if fdSet.IsSet(abortFd) {
		return errAborted
	}

	if fdSet.IsSet(readerFd) {
		return nil
	}

	return fmt.Errorf("select returned without setting a file descriptor")
}
