package progress

import (
	"fmt"
	"time"
)

// Elapsed says how long something has been going on: "12s", "1m 05s". Under
// a second it says nothing, so a wait that is over before it is noticed adds
// no number.
func Elapsed(d time.Duration) string {
	s := int(d.Seconds())
	switch {
	case s < 1:
		return ""
	case s < 60:
		return fmt.Sprintf("%ds", s)
	}
	return fmt.Sprintf("%dm %02ds", s/60, s%60)
}

// OfProjects names the projects a wait works on: "every project" while their
// number is not known, else "the 23 projects" ("the project" for one).
func OfProjects(n int) string {
	switch {
	case n <= 0:
		return "every project"
	case n == 1:
		return "the project"
	}
	return fmt.Sprintf("the %d projects", n)
}

// Waiting appends the elapsed time to what a wait is doing, once it is long
// enough to notice: "Asking npm… 12s".
func Waiting(what string, d time.Duration) string {
	if d < 2*time.Second {
		return what
	}
	return what + " " + Elapsed(d)
}
