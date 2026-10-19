package progress

// Filled is how many of width cells are filled for done out of total. A bar
// is never full until everything is done, and never empty once something is.
func Filled(done, total, width int) int {
	if width <= 0 || total <= 0 || done <= 0 {
		return 0
	}
	if done >= total {
		return width
	}
	n := done * width / total
	return min(max(n, 1), width-1)
}

// Percent is the whole-number percentage done, from 0 to 100.
func Percent(done, total int) int {
	if total <= 0 || done <= 0 {
		return 0
	}
	return min(done*100/total, 100)
}
