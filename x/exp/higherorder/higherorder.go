package higherorder

// Foldl applies a function to each element of a list, starting from the left.
// A single value is returned.
func Foldl[A any](f func(x, y A) A, start A, list []A) A {
	for _, v := range list {
		start = f(start, v)
	}
	return start
}

// Foldr applies a function to each element of a list, starting from the right.
// A single value is returned.
func Foldr[A any](f func(x, y A) A, start A, list []A) A {
	for i := len(list) - 1; i >= 0; i-- {
		start = f(start, list[i])
	}
	return start
}

// Map applies a given function to each element of a list, returning a new list.
func Map[A, B any](f func(A) B, list []A) []B {
	res := make([]B, len(list))
	for i, v := range list {
		res[i] = f(v)
	}
	return res
}
