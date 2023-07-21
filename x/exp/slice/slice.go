package slice

// Take returns the first n elements of the given slice. If there are not
// enough elements in the slice, the whole slice is returned.
func Take[A any](slice []A, n int) []A {
	if n > len(slice) {
		return slice
	}
	return slice[:n]
}
