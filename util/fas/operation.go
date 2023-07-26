package fas

// TernaryOp is ternary operation like max = a > b ? a : b
func TernaryOp[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}
