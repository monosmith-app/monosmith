package verifyfacts

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// passedN and failedN read "4 passed" and "1 failed" (jest, vitest, pytest,
	// cargo) and dotnet's "Passed: 4" and "Failed: 1".
	passedN = regexp.MustCompile(`(?i)\b(\d+) passed\b|\bPassed:\s*(\d+)`)
	failedN = regexp.MustCompile(`(?i)\b(\d+) failed\b|\bFailed:\s*(\d+)`)
	// cargoResult is the line each test binary of cargo test ends with.
	cargoResult = regexp.MustCompile(`^test result: (?:ok|FAILED)\.`)
	// goOK and goFail are the lines go test prints for a package that passed and
	// for a package that failed.
	goOK   = regexp.MustCompile(`^ok\s+\S+`)
	goFail = regexp.MustCompile(`^FAIL\s+\S+`)
	// fileSummary is a line that counts files or suites, not tests: it is skipped.
	fileSummary = regexp.MustCompile(`^\s*(Test Files|Test Suites)\b`)
)

// TestCounter reads a test task's output line by line and keeps what the runner
// said about the tests. It knows the summary lines of jest, vitest, pytest,
// dotnet test and cargo test (which prints one per test binary, so they add up),
// and for go test, which prints a line per package, the packages that passed and
// failed. Anything else leaves the counts unknown.
type TestCounter struct {
	passed, failed int
	// cargoPassed and cargoFailed add the lines of every cargo test binary.
	cargoPassed, cargoFailed int
	cargoSeen                bool
	// packages and failedPackages are go test's.
	packages, failedPackages int
}

// Add reads one line of output, without its newline and colour.
func (c *TestCounter) Add(line string) {
	switch {
	case goOK.MatchString(line):
		c.packages++
		return
	case goFail.MatchString(line):
		c.failedPackages++
		return
	case fileSummary.MatchString(line):
		return
	case cargoResult.MatchString(line):
		c.cargoSeen = true
		c.cargoPassed += first(passedN, line)
		c.cargoFailed += first(failedN, line)
		return
	}
	// The last summary wins: a runner prints the files before the tests.
	if n, ok := count(passedN, line); ok {
		c.passed = n
	}
	if n, ok := count(failedN, line); ok {
		c.failed = n
	}
}

// Result is the tests the runner reported. Known is false when no line said anything.
//
// Returns: the counts, what they count ("tests" or "packages") and whether any were read.
func (c TestCounter) Result() (passed, failed int, unit string, known bool) {
	switch {
	case c.cargoSeen:
		return c.cargoPassed, c.cargoFailed, "tests", true
	case c.passed > 0 || c.failed > 0:
		return c.passed, c.failed, "tests", true
	case c.packages > 0 || c.failedPackages > 0:
		return c.packages, c.failedPackages, "packages", true
	}
	return 0, 0, "", false
}

// count is the number the first group of the pattern captured, when it matched.
func count(re *regexp.Regexp, line string) (int, bool) {
	m := re.FindStringSubmatch(line)
	if m == nil {
		return 0, false
	}
	digits := m[1]
	if digits == "" {
		digits = m[2]
	}
	n, err := strconv.Atoi(strings.TrimSpace(digits))
	return n, err == nil
}

func first(re *regexp.Regexp, line string) int {
	n, _ := count(re, line)
	return n
}
