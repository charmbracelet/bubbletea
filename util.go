package tea

import "runtime"

// isWindows return whether the current runtime is Windows.
func isWindows() bool {
	return runtime.GOOS == "windows"
}
