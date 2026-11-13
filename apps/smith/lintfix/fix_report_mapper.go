package lintfix

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
)

// shownRemedies caps the "→ see them" lines.
const shownRemedies = 5

// LintCommand is the command that shows a project's lint problems.
//
// Parameters: project is the project's name. Returns the command line.
func LintCommand(project string) string {
	return "smith run lint " + flagvocabulary.Arg("project") + " " + project
}

// rank orders the rows: a tool that could not run, then problems left, then
// the projects not reached, then the clean ones.
func rank(o Outcome) int {
	switch {
	case o.Crash != "":
		return 0
	case o.Problems() > 0:
		return 1
	case o.Skipped:
		return 2
	}
	return 3
}

// Report writes what `smith lint --fix` did: one table with a row per project
// (failures first, then the rest alphabetically) with the problems fixed and
// left, the cause and remedy of each tool that could not run, one line naming
// the projects whose tool has no fix mode, and a closing line with the totals
// and the time.
//
// Parameters: outcomes are the projects' results; took is the run's duration.
// Returns the text, and failed: true when a tool could not run or a problem is
// left.
func Report(outcomes []Outcome, took time.Duration) (text string, failed bool) {
	rows := append([]Outcome(nil), outcomes...)
	sort.SliceStable(rows, func(i, j int) bool {
		if rank(rows[i]) != rank(rows[j]) {
			return rank(rows[i]) < rank(rows[j])
		}
		return rows[i].Project < rows[j].Project
	})
	width := len("Project")
	for _, o := range rows {
		width = max(width, len(o.Project))
	}
	var b strings.Builder
	line := func(project, fixed, left, result string) {
		fmt.Fprintf(&b, "%-*s  %6s  %6s  %s\n", width, project, fixed, left, result)
	}
	line("Project", "Fixed", "Left", "Result")
	var crashed, left []Outcome
	var noFix []string
	fixed, total, skipped := 0, 0, 0
	for _, o := range rows {
		if o.NoFix {
			noFix = append(noFix, o.Project)
		}
		fixedCell := "-"
		if o.Fixed >= 0 && !o.NoFix {
			fixedCell = strconv.Itoa(o.Fixed)
		}
		switch rank(o) {
		case 0:
			crashed = append(crashed, o)
			line(o.Project, "-", "-", marks.Failed+" could not run: "+o.Crash)
			continue
		case 1:
			left = append(left, o)
			leftCell, result := "some", marks.Warning+" problems left"
			if o.Counted {
				leftCell = strconv.Itoa(o.Problems())
				result = fmt.Sprintf("%s %s, %s", marks.Warning, count(o.Errors, "error"), count(o.Warnings, "warning"))
			}
			line(o.Project, fixedCell, leftCell, result)
		case 2:
			skipped++
			line(o.Project, "-", "-", marks.Skipped+" not run: a task it needs failed")
			continue
		default:
			line(o.Project, fixedCell, "0", marks.Passed+" no problems left")
		}
		fixed += max(o.Fixed, 0)
		total += o.Problems()
	}
	for _, o := range crashed {
		fmt.Fprintf(&b, "%s %s: could not run: %s\n→ %s\n", marks.Failed, o.Project, o.Crash, o.Remedy)
	}
	if len(noFix) > 0 {
		fmt.Fprintf(&b, "%s No fix mode for %s: the plain lint ran, nothing was changed\n", marks.Warning, strings.Join(noFix, ", "))
	}
	failed = len(crashed) > 0 || len(left) > 0 || skipped > 0
	mark := marks.Passed
	switch {
	case len(crashed) > 0:
		mark = marks.Failed
	case failed:
		mark = marks.Warning
	}
	verdict := "passed"
	if failed {
		verdict = "found problems"
	}
	leftText := "none left"
	if total > 0 {
		leftText = fmt.Sprintf("%d left in %s", total, count(len(left), "project"))
	}
	fmt.Fprintf(&b, "%s Lint fix %s in %s: %s, %s fixed, %s\n", mark, verdict, took.Round(time.Second), count(len(rows), "project"), count(fixed, "problem"), leftText)
	for i, o := range left {
		if i == shownRemedies {
			fmt.Fprintf(&b, "→ and %d more projects\n", len(left)-shownRemedies)
			break
		}
		fmt.Fprintf(&b, "→ see them: %s\n", LintCommand(o.Project))
	}
	return b.String(), failed
}

func count(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
