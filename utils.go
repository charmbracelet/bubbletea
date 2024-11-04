package tea

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(x, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}
