package cipipeline

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/verifyfacts"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// detailsReport is what verify says about each project beyond pass or fail: the
// tests, the lint issues, the size of the build and the coverage in one table,
// and where the build, the test results and the coverage report are in another.
type detailsReport struct {
	// Facts is the first table: Project, Tests, Lint, Build, Coverage.
	Facts resultTable
	// Paths is the second: Project, Build, Test results, Coverage report.
	Paths resultTable
	// CoverageNote is a line that says no project has a coverage report in place, or "".
	CoverageNote string
}

// reportDetails builds the tables of a verify run. A number smith could not read
// is left as "-" and never guessed. Coverage and test results are shown only when a
// report is already in place: smith never turns them on.
//
// Parameters: results are the task results of the run; outputs hold what each
// task printed; projects are the workspace's projects; root is the workspace
// folder; since is when the run began (a report written before it shows its age);
// icons are the status glyphs.
// Returns: the tables; empty when no project ran a test, lint or build task.
func reportDetails(results []taskrun.Result, outputs *taskOutputs, projects []*projectgraph.Project, root string, since time.Time, icons theme.Icons) detailsReport {
	status := map[string]taskrun.Result{}
	for _, r := range results {
		status[r.Task] = r
	}
	byName := map[string]*projectgraph.Project{}
	for _, p := range projects {
		byName[p.Name] = p
	}
	names := map[string]bool{}
	for _, r := range results {
		project, task := projectOf(r.Task)
		if task == "test" || task == "lint" || task == "build" {
			names[project] = true
		}
	}
	sorted := make([]string, 0, len(names))
	for n := range names {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)

	rep := detailsReport{}
	rep.Facts.Headers = []string{"Project", "Tests", "Lint", "Build", "Coverage"}
	rep.Facts.Right = []bool{false, false, false, true, true}
	rep.Paths.Headers = []string{"Project", "Build path", "Test results", "Coverage report"}
	anyTests, anyCoverage := false, false
	var bad, good []tableRow
	var paths []tableRow
	for _, name := range sorted {
		p := byName[name]
		row := tableRow{Cells: []string{name}}
		pathRow := tableRow{Cells: []string{name, "-", "-", "-"}}
		if r, ok := status[name+":test"]; ok {
			anyTests = true
			cell, isBad := testsCell(r, outputs, name+":test", icons)
			row.Cells, row.Bad = append(row.Cells, cell), row.Bad || isBad
		} else {
			row.Cells = append(row.Cells, icons.Queued+" none")
		}
		if r, ok := status[name+":lint"]; ok {
			cell, isBad := lintCell(r, outputs, name+":lint", icons)
			row.Cells, row.Bad = append(row.Cells, cell), row.Bad || isBad
		} else {
			row.Cells = append(row.Cells, icons.Queued+" none")
		}
		buildCell, coverageCell := "-", "-"
		if p != nil {
			if r, ok := status[name+":build"]; ok && (r.Status == taskrun.Succeeded || r.Status == taskrun.Cached) {
				if b, found := verifyfacts.FindBuild(root, protocol.BuildOutputDir(p.Project)); found {
					buildCell, pathRow.Cells[1] = verifyfacts.FormatSize(b.Bytes), b.Path
				}
			}
			if results, found := verifyfacts.FindResults(root, p.Root); found {
				pathRow.Cells[2] = results
			}
			if c, found := verifyfacts.FindCoverage(root, p.Root); found {
				anyCoverage = true
				coverageCell, pathRow.Cells[3] = coverageText(c, since), c.Path
			}
		}
		row.Cells = append(row.Cells, buildCell, coverageCell)
		if row.Bad {
			bad = append(bad, row)
		} else {
			good = append(good, row)
		}
		if pathRow.Cells[1] != "-" || pathRow.Cells[2] != "-" || pathRow.Cells[3] != "-" {
			paths = append(paths, pathRow)
		}
	}
	rep.Facts.Rows = append(bad, good...)
	rep.Paths.Rows = paths
	if anyTests && !anyCoverage {
		rep.CoverageNote = "Coverage: not collected. smith shows it when a report is already in place (coverage/coverage-summary.json, lcov.info, coverage.xml, a Go coverage profile)."
	}
	return rep
}

// empty is true when there is nothing to show.
func (d detailsReport) empty() bool { return len(d.Facts.Rows) == 0 }

// testsCell is the Tests cell: the status glyph, then what the runner said.
func testsCell(r taskrun.Result, outputs *taskOutputs, id string, icons theme.Icons) (cell string, bad bool) {
	passed, failed, unit, known := outputs.TestFacts(id)
	switch r.Status {
	case taskrun.Failed:
		noun := strings.TrimSuffix(unit, "s")
		switch {
		case known && failed > 0 && passed > 0:
			return fmt.Sprintf("%s %d failed, %d passed", icons.Failed, failed, passed), true
		case known && failed > 0:
			return fmt.Sprintf("%s %s failed", icons.Failed, plural(failed, noun)), true
		}
		return icons.Failed + " failed", true
	case taskrun.Skipped:
		return icons.Skipped + " skipped", true
	}
	glyph := icons.Passed
	if r.Status == taskrun.Cached {
		glyph = icons.Cached
	}
	if known {
		return fmt.Sprintf("%s %s passed", glyph, plural(passed, strings.TrimSuffix(unit, "s"))), false
	}
	return glyph, false
}

// lintCell is the Lint cell: "✓ passed", or the number of issues the linter reported.
func lintCell(r taskrun.Result, outputs *taskOutputs, id string, icons theme.Icons) (cell string, bad bool) {
	issues, known := outputs.LintIssues(id)
	switch r.Status {
	case taskrun.Failed:
		if known && issues > 0 {
			return fmt.Sprintf("%s %s", icons.Failed, plural(issues, "issue")), true
		}
		return icons.Failed + " failed", true
	case taskrun.Skipped:
		return icons.Skipped + " skipped", true
	}
	glyph := icons.Passed
	if r.Status == taskrun.Cached {
		glyph = icons.Cached
	}
	if known && issues > 0 {
		// The linter reported issues and still passed (warnings).
		return fmt.Sprintf("%s %s", icons.Warning, plural(issues, "warning")), false
	}
	return glyph + " passed", false
}

// coverageText is "81.2%", with the age of a report written before the run began.
func coverageText(c verifyfacts.Coverage, since time.Time) string {
	text := fmt.Sprintf("%.1f%%", c.Percent)
	if !since.IsZero() && c.Modified.Before(since) {
		text += " (" + ageText(since.Sub(c.Modified)) + " old)"
	}
	return text
}

// ageText words a span as "2h", "3d" or "5m".
func ageText(d time.Duration) string {
	switch {
	case d >= 24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d >= time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dm", max(int(d.Minutes()), 1))
}
