package projectstandard

import "fmt"

// The answers to "your project configuration differs from smith's standard".
const (
	// Replace writes smith's standard files; a whole file replaced is kept as
	// <file>.pre-smith. It is the default.
	Replace = "replace"
	// Keep leaves the project's files as they are; smith doctor still says
	// where they differ, until the warning is accepted.
	Keep = "keep"
)

// ParseAnswer reads --project-config: "" (not given), replace or keep.
//
// Errors: any other value.
func ParseAnswer(v string) (string, error) {
	switch v {
	case "", Replace, Keep:
		return v, nil
	}
	return "", fmt.Errorf("--project-config %q: say replace (smith's standard files; yours kept as <file>.pre-smith) or keep", v)
}
