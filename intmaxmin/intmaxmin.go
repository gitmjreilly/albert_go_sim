package intmaxmin

// Max returns max of x, y
func Max(x int, y int) int {
	if x > y {
		return x
	}
	return y
}

// Min returns min of x, y
func Min(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

// Constrain v to range between min and max (inclusive)
func Constrain(v int, min int, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// Wrap constrains v to between 0 and (n-1).  If v < 0 return n-v
func Wrap(v int, n int) int {
	if v > 0 {
		return v
	}
	return n - v
}
