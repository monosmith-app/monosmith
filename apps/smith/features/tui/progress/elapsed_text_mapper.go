package progress

import (
	"fmt"
	"time"

	"github.com/monosmith-app/monosmith/libs/words/words"
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
	return words.Say("progress.elapsed.minutes", "minutes", s/60, "seconds", fmt.Sprintf("%02d", s%60))
}

// OfProjects names the projects a wait works on: "every project" while their
// number is not known, else "the 23 projects" ("the project" for one).
func OfProjects(n int) string {
	switch {
	case n <= 0:
		return words.Say("progress.projects.every")
	case n == 1:
		return words.Say("progress.projects.one")
	}
	return words.Say("progress.projects.many", "count", n)
}

// Waiting appends the elapsed time to what a wait is doing, once it is long
// enough to notice: "Asking npm… 12s".
func Waiting(what string, d time.Duration) string {
	if d < 2*time.Second {
		return what
	}
	return what + " " + Elapsed(d)
}
