package cipipeline

import (
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// stepVerdict is how a plugin step ended.
type stepVerdict int

const (
	stepRan stepVerdict = iota
	// stepUnchanged: an audit verdict from an earlier run still holds.
	stepUnchanged
	// stepProblem: a non-blocking step reported problems.
	stepProblem
	// stepSkipped: a non-blocking step did not run because its tool is missing.
	stepSkipped
)

// stepOutcome is one plugin step's end, for the table after the folded output.
type stepOutcome struct {
	Name    string
	Verdict stepVerdict
	Took    time.Duration
	// Since is the day an unchanged step last passed (YYYY-MM-DD).
	Since string
}

// stepTable is one row per step, in run order. The Note column exists only
// when some step has something to add.
func stepTable(kind string, outcomes []stepOutcome, icons theme.Icons) resultTable {
	t := resultTable{Headers: []string{kind, "Result", "Time"}, Right: []bool{false, false, true}}
	notes := false
	for _, o := range outcomes {
		row := tableRow{Cells: []string{o.Name, icons.Passed + " done", formatDuration(o.Took)}}
		switch o.Verdict {
		case stepUnchanged:
			row.Cells[1], row.Cells[2] = icons.Skipped+" unchanged", "-"
			row.Cells = append(row.Cells, "nothing changed since it passed on "+o.Since)
			notes = true
		case stepProblem:
			row.Cells[1], row.Bad = icons.Warning+" problems", true
			row.Cells = append(row.Cells, "reported problems, not blocking")
			notes = true
		case stepSkipped:
			row.Cells[1], row.Cells[2], row.Bad = icons.Warning+" skipped", "-", true
			row.Cells = append(row.Cells, "a tool it needs is not installed: nothing was checked")
			notes = true
		}
		t.Rows = append(t.Rows, row)
	}
	if notes {
		t.Headers = append(t.Headers, "Note")
		for i := range t.Rows {
			if len(t.Rows[i].Cells) == 3 {
				t.Rows[i].Cells = append(t.Rows[i].Cells, "")
			}
		}
	}
	return t
}

// countOutcomes counts the steps by verdict.
func countOutcomes(outcomes []stepOutcome) (ran, unchanged, problems, skipped int) {
	for _, o := range outcomes {
		switch o.Verdict {
		case stepUnchanged:
			unchanged++
		case stepProblem:
			problems++
		case stepSkipped:
			skipped++
		case stepRan:
			ran++
		}
	}
	return ran, unchanged, problems, skipped
}

// problemNames lists the steps that reported problems.
func problemNames(outcomes []stepOutcome) []string {
	var names []string
	for _, o := range outcomes {
		if o.Verdict == stepProblem {
			names = append(names, o.Name)
		}
	}
	return names
}

// listSome names the first few and counts the rest: "a, b, c and 10 more".
func listSome(names []string, few int) string {
	if len(names) <= few {
		return strings.Join(names, ", ")
	}
	return fmt.Sprintf("%s and %d more", strings.Join(names[:few], ", "), len(names)-few)
}

// setupConclusion is the end of the setup's closing line.
func setupConclusion(outcomes []stepOutcome, projects int) string {
	return fmt.Sprintf("%s ran, %s ready", plural(len(outcomes), "step"), plural(projects, "project"))
}

// auditConclusion is the end of the audit's closing line.
func auditConclusion(outcomes []stepOutcome, projects int) string {
	ran, unchanged, problems, skipped := countOutcomes(outcomes)
	parts := []string{fmt.Sprintf("%d ran", ran)}
	if unchanged > 0 {
		parts = append(parts, fmt.Sprintf("%d unchanged since an earlier pass", unchanged))
	}
	if problems > 0 {
		parts = append(parts, fmt.Sprintf("%d reported problems (not blocking)", problems))
	}
	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped because a tool is missing (not audited)", skipped))
	}
	return fmt.Sprintf("%s over %s: %s", plural(len(outcomes), "check"), plural(projects, "project"), strings.Join(parts, ", "))
}
