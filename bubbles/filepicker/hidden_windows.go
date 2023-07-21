//go:build windows
// +build windows

package filepicker

import (
	"syscall"
)

// IsHidden reports whether a file is hidden or not.
func IsHidden(file string) (bool, error) {
	pointer, err := syscall.UTF16PtrFromString(file)
	if err != nil {
		return false, err
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}
	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
