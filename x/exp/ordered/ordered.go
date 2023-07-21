package ordered

import (
	"golang.org/x/exp/constraints"
)

// Max returns the smaller of a and b.
func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of a and b.
func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Clamp returns a value clamped between the given low and high values.
func Clamp[T constraints.Ordered](n, low, high T) T {
	if low > high {
		low, high = high, low
	}
	return Min(high, Max(low, n))
}
