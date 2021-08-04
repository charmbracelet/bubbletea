// +build windows

package tea

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
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

	return nil
}

var (
	winsocks2 = syscall.NewLazyDLL("ws2_32.dll")

	selectSyscall = winsocks2.NewProc("select")
	isSetSyscall  = winsocks2.NewProc("__WSAFDIsSet")
)

func _select(nfds int, reader *os.File, abort *os.File) error {
	readerFd := reader.Fd()
	abortFd := abort.Fd()

	maxFd := readerFd
	if abortFd > maxFd {
		maxFd = abortFd
	}

	if maxFd >= 1024 {
		return fmt.Errorf("cannot select on file descriptor %d which is larger than 1024", maxFd)
	}

	fdSet := &fdSet{}
	fdSet.Set(readerFd)
	fdSet.Set(abortFd)

	res, _, errno := syscall.Syscall6(selectSyscall.Addr(), 5, uintptr(nfds),
		uintptr(unsafe.Pointer(fdSet)), 0, 0, 0, 0)
	if int(res) == 0 {
		if errno == 0 {
			return error(syscall.EINVAL)
		}

		return fmt.Errorf("select: %w", error(errno))
	}

	aborted, err := isSet(fdSet, abortFd)
	if err != nil {
		return errno
	}

	if aborted {
		return errAborted
	}

	ready, err := isSet(fdSet, readerFd)
	if err != nil {
		return err
	}

	if !ready {
		return fmt.Errorf("select returned without setting a file descriptor")
	}

	return nil
}

func isSet(fdset *fdSet, fd uintptr) (bool, error) {
	res, _, errno := syscall.Syscall(isSetSyscall.Addr(), 2, uintptr(syscall.Handle(fd)),
		uintptr(unsafe.Pointer(fdset)), 0)
	if int(res) == 0 {
		if errno == 0 {
			return false, error(syscall.EINVAL)
		}

		return false, fmt.Errorf("checking if file descriptor is set: %w", error(errno))
	}

	return int(res) != 0, nil
}

const fdSetSize = 64

type fdSet struct {
	count uint
	fds   [fdSetSize]uintptr
}

// Set adds the fd to the set
func (fds *fdSet) Set(fd uintptr) {
	var i uint
	for i = 0; i < fds.count; i++ {
		if fds.fds[i] == fd {
			break
		}
	}
	if i == fds.count {
		if fds.count < fdSetSize {
			fds.fds[i] = fd
			fds.count++
		}
	}
}
