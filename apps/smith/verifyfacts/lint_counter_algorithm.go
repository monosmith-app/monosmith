package verifyfacts

import (
	"regexp"
	"strconv"
)

var (
	// eslintTotal reads the line eslint ends with: a heavy cross, then "3 problems (2
	// errors, 1 warning)". The cross is written as its code point, because the
	// glyphs of smith come from the marks package.
	eslintTotal = regexp.MustCompile(`[\x{2716}x\x{00D7}]\s+(\d+) problems?\b`)
	// golangciTotal reads "5 issues:" and "0 issues.".
	golangciTotal = regexp.MustCompile(`^(\d+) issues?[:.]`)
	// ruffTotal reads "Found 3 errors." and "Found 1 error.".
	ruffTotal = regexp.MustCompile(`^Found (\d+) errors?\b`)
	// ruffClean is what ruff prints when nothing was found.
	ruffClean = regexp.MustCompile(`^All checks passed!`)
)

// LintCounter reads a lint task's output line by line and keeps the number of
// issues the linter reported. It knows eslint, golangci-lint and ruff; for any
// other linter the count stays unknown and the status of the task is all there is.
type LintCounter struct {
	issues int
	seen   bool
}

// Add reads one line of output, without its newline and colour.
func (c *LintCounter) Add(line string) {
	for _, re := range []*regexp.Regexp{eslintTotal, golangciTotal, ruffTotal} {
		if m := re.FindStringSubmatch(line); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil {
				c.issues, c.seen = n, true
			}
			return
		}
	}
	if ruffClean.MatchString(line) {
		c.issues, c.seen = 0, true
	}
}

// Issues is the number of issues the linter reported; known is false when its
// output said nothing smith can read.
func (c LintCounter) Issues() (issues int, known bool) { return c.issues, c.seen }
