package cipipeline

import (
	"regexp"
	"strconv"
)

var (
	// passedCount matches "42 passed" (pytest, vitest, jest) and dotnet's
	// "Passed: 42".
	passedCount = regexp.MustCompile(`(?i)\b(\d+) passed\b|\bPassed:\s*(\d+)`)
	// goPackageOK matches the line `go test` prints for a passing package.
	goPackageOK = regexp.MustCompile(`^ok\s+\S+`)
)

// testCounter reads a test task's output line by line and keeps what the
// runner said about how many tests passed. It knows the summary lines of
// pytest, vitest, jest and dotnet test; for go test, which prints one line per
// package, it counts the packages. Anything else leaves the count unknown.
type testCounter struct {
	passed   int
	packages int
}

// add reads one line of output.
func (c *testCounter) add(line string) {
	if goPackageOK.MatchString(line) {
		c.packages++
		return
	}
	if m := passedCount.FindStringSubmatch(line); m != nil {
		// The last summary wins: vitest and jest print the files before the tests.
		digits := m[1]
		if digits == "" {
			digits = m[2]
		}
		if n, err := strconv.Atoi(digits); err == nil {
			c.passed = n
		}
	}
}

// summary is "42 passed", "12 packages ok" or "" when the runner said nothing
// smith can read.
func (c testCounter) summary() string {
	switch {
	case c.passed > 0:
		return strconv.Itoa(c.passed) + " passed"
	case c.packages > 0:
		return plural(c.packages, "package") + " ok"
	}
	return ""
}
