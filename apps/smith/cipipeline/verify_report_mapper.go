package cipipeline

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// columnTitles name the verify tasks in a table header.
var columnTitles = map[string]string{"lint": "Lint", "typecheck": "Types", "test": "Tests", "build": "Build", "package": "Package"}

// projectOf splits a task id ("core:test") into its project and task.
func projectOf(id string) (project, task string) {
	i := strings.LastIndex(id, ":")
	if i < 0 {
		return id, ""
	}
	return id[:i], id[i+1:]
}

// taskReport is the table of one run of tasks: a row per project, a column per
// task, and what the closing line says.
type taskReport struct {
	Table      resultTable
	Conclusion string
	// Failed lists the failed or skipped task ids, in table order.
	Failed []string
	// Projects is how many projects ran a task.
	Projects int
}

// reportTasks turns results into a table of projects by tasks. columns are the
// tasks to show, in order; a project without a task shows "○ none". Failures
// come first, then the projects alphabetically. outputs gives each test run's
// own summary ("42 passed").
func reportTasks(results []taskrun.Result, columns []string, outputs *taskOutputs, icons theme.Icons) taskReport {
	type cell struct {
		text   string
		bad    bool
		took   time.Duration
		cached bool
	}
	byProject := map[string]map[string]cell{}
	tests, cachedCount, ok := 0, 0, 0
	var failed, skipped []string
	for _, r := range results {
		project, task := projectOf(r.Task)
		if byProject[project] == nil {
			byProject[project] = map[string]cell{}
		}
		c := cell{took: r.Duration}
		switch r.Status {
		case taskrun.Failed:
			c.text, c.bad = icons.Failed+" failed", true
			failed = append(failed, r.Task)
		case taskrun.Skipped:
			c.text, c.bad = icons.Skipped+" skipped", true
			skipped = append(skipped, r.Task)
		case taskrun.Cached:
			c.text, c.cached = icons.Cached+" cached", true
			cachedCount++
			ok++
		default:
			c.text = icons.Passed
			ok++
		}
		if task == "test" && (r.Status == taskrun.Succeeded || r.Status == taskrun.Cached) {
			if summary, n := outputs.Tests(r.Task); summary != "" {
				if c.cached {
					c.text = icons.Cached
				}
				c.text += " " + summary
				tests += n
			}
		}
		byProject[project][task] = c
	}
	names := make([]string, 0, len(byProject))
	for p := range byProject {
		names = append(names, p)
	}
	sort.Strings(names)
	rep := taskReport{Projects: len(names)}
	rep.Table.Headers = append([]string{"Project"}, titles(columns)...)
	rep.Table.Headers = append(rep.Table.Headers, "Time")
	rep.Table.Right = make([]bool, len(rep.Table.Headers))
	rep.Table.Right[len(rep.Table.Right)-1] = true
	var bad, good []tableRow
	for _, p := range names {
		row := tableRow{Cells: []string{p}}
		var total time.Duration
		for _, col := range columns {
			c, has := byProject[p][col]
			if !has {
				row.Cells = append(row.Cells, icons.Queued+" none")
				continue
			}
			row.Cells = append(row.Cells, c.text)
			row.Bad = row.Bad || c.bad
			total += c.took
		}
		if total > 0 {
			row.Cells = append(row.Cells, formatDuration(total))
		} else {
			row.Cells = append(row.Cells, "-")
		}
		if row.Bad {
			bad = append(bad, row)
		} else {
			good = append(good, row)
		}
	}
	rep.Table.Rows = append(bad, good...)
	rep.Failed = append(failed, skipped...)
	parts := []string{plural(len(names), "project") + " checked", fmt.Sprintf("%d of %s ok", ok, plural(len(results), "task"))}
	if cachedCount > 0 {
		parts[1] += fmt.Sprintf(" (%d from cache)", cachedCount)
	}
	if tests > 0 {
		parts = append(parts, fmt.Sprintf("%d tests passed", tests))
	}
	rep.Conclusion = strings.Join(parts, ", ")
	return rep
}

func titles(columns []string) []string {
	out := make([]string, len(columns))
	for i, c := range columns {
		if t, ok := columnTitles[c]; ok {
			out[i] = t
		} else {
			out[i] = strings.ToUpper(c[:1]) + c[1:]
		}
	}
	return out
}

// failureLines are the way forward after failed tasks: the end of the first
// few failures' output, then the command that repeats the first.
func failureLines(failed []string, outputs *taskOutputs) []string {
	const shownFailures, shownLines = 3, 8
	var lines []string
	for i, id := range failed {
		if i == shownFailures {
			lines = append(lines, fmt.Sprintf("... and %d more; the folded output above has them all", len(failed)-shownFailures))
			break
		}
		if tail := outputs.Tail(id); tail != "" {
			lines = append(lines, "last output of "+id+":")
			rows := strings.Split(tail, "\n")
			for _, l := range rows[max(len(rows)-shownLines, 0):] {
				lines = append(lines, "  "+l)
			}
		}
	}
	if len(failed) > 0 {
		project, task := projectOf(failed[0])
		lines = append(lines, fmt.Sprintf("→ repeat it with `smith run %s -p %s`", task, project))
	}
	return lines
}
