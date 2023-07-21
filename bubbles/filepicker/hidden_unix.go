//go:build !windows
// +build !windows

package filepicker

import "strings"

// IsHidden reports whether a file is hidden or not.
func IsHidden(file string) (bool, error) {
	return strings.HasPrefix(file, "."), nil
}
