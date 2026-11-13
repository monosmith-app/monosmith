package lintfix

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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

// Report writes what ESLint did: a row per project (crashes first, then the
// projects with problems left, then the clean ones, each alphabetically) with
// the problems fixed and left, then closing lines with the totals and the
// remedy.
//
// Parameters: outcomes are Fix's results. Returns the text, and failed: true
// when ESLint crashed somewhere or left a problem.
func Report(outcomes []Outcome) (text string, failed bool) {
	rows := append([]Outcome(nil), outcomes...)
	rank := func(o Outcome) int {
		switch {
		case o.Crash != "":
			return 0
		case o.Problems() > 0:
			return 1
		}
		return 2
	}
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
	fixed, total := 0, 0
	for _, o := range rows {
		switch rank(o) {
		case 0:
			crashed = append(crashed, o)
			line(o.Project, "-", "-", marks.Failed+" ESLint could not run: "+o.Crash)
			continue
		case 1:
			left = append(left, o)
			line(o.Project, strconv.Itoa(o.Fixed), strconv.Itoa(o.Problems()), fmt.Sprintf("%s %s, %s", marks.Warning, count(o.Errors, "error"), count(o.Warnings, "warning")))
		default:
			line(o.Project, strconv.Itoa(o.Fixed), "0", marks.Passed+" no problems left")
		}
		fixed += o.Fixed
		total += o.Problems()
	}
	for _, o := range crashed {
		fmt.Fprintf(&b, "%s %s: ESLint could not run: %s\n→ %s\n", marks.Failed, o.Project, o.Crash, o.Remedy)
	}
	ran := len(rows) - len(crashed)
	switch {
	case ran == 0:
	case len(left) > 0:
		fmt.Fprintf(&b, "%s ESLint fixed %s; %d %s left in %d of %s\n", marks.Warning, count(fixed, "problem"), total, isAre(total), len(left), count(ran, "project"))
		for i, o := range left {
			if i == shownRemedies {
				fmt.Fprintf(&b, "→ and %d more projects\n", len(left)-shownRemedies)
				break
			}
			fmt.Fprintf(&b, "→ see them: %s\n", LintCommand(o.Project))
		}
	default:
		fmt.Fprintf(&b, "%s ESLint fixed %s; none are left in %s\n", marks.Passed, count(fixed, "problem"), count(ran, "project"))
	}
	return b.String(), len(crashed) > 0 || len(left) > 0
}

// count spells a number with its noun: "1 problem", "12 problems".
func count(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func isAre(n int) string {
	if n == 1 {
		return "is"
	}
	return "are"
}
