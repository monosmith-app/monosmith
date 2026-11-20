package projecttooling

// Answer is what a project uses for its build and lint.
type Answer string

// The answers.
const (
	// Keep leaves the project's own build and lint tools as they are.
	Keep Answer = "keep"
	// Smith puts smith's build and ESLint config in place, then retires the
	// old tools.
	Smith Answer = "smith"
)
