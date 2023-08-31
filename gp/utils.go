package gp

// todo use slices when go is 1.20
func ReverseSlice[T comparable](s []T) []T {
	var r []T
	for i := len(s) - 1; i >= 0; i-- {
		r = append(r, s[i])
	}
	return r
}

func Max(a int, b int) int {
	if a >= b {
		return a
	}
	return b
}

func Append(s []int, times int, value int) []int {
	for i := 0; i < times; i++ {
		s = append(s, value)
	}
	return s
}

func Pop[T any](s []T) ([]T, T) {
	return s[:len(s)-1], s[len(s)-1]
}
