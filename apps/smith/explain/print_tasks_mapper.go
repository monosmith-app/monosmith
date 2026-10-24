package explain

import (
	"fmt"
	"io"
	"strings"
)

// printTasks writes tasks for a person: a one-line summary of each task
// first, then each task's fields, with where a value came from said once
// where it is the same for all of them.
func printTasks(w io.Writer, tasks []Task) {
	if len(tasks) == 0 {
		return
	}
	common := commonSource(tasks)
	if len(tasks) > 1 {
		fmt.Fprintf(w, "%s has %d tasks:\n", project(tasks[0].ID), len(tasks))
		width := 0
		for _, t := range tasks {
			width = max(width, len(taskName(t.ID)))
		}
		for _, t := range tasks {
			fmt.Fprintf(w, "  %-*s  %s\n", width, taskName(t.ID), summary(t))
		}
		fmt.Fprintln(w)
	}
	for _, t := range tasks {
		fmt.Fprintf(w, "%s", t.ID)
		if t.Description != "" {
			fmt.Fprintf(w, " — %s", t.Description)
		}
		fmt.Fprintln(w)
		for _, f := range t.Fields {
			fmt.Fprintf(w, "  %-10s %s", f.Name, display(f))
			if f.Source != "" && f.Source != common {
				fmt.Fprintf(w, "   (from %s)", f.Source)
			}
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w)
	}
	if common != "" {
		fmt.Fprintf(w, "Every value above comes from %s, except where marked.\n", common)
	}
}

// commonSource is the source most fields share, when every field of every
// task names one; "" when they differ so much that each needs its label.
func commonSource(tasks []Task) string {
	count := map[string]int{}
	total := 0
	for _, t := range tasks {
		for _, f := range t.Fields {
			if f.Source != "" {
				count[f.Source]++
				total++
			}
		}
	}
	best, n := "", 0
	for source, c := range count {
		if c > n {
			best, n = source, c
		}
	}
	if total == 0 || n*2 < total {
		return ""
	}
	return best
}

// summary says in a line what a task does: its command, and whether its
// result is cached.
func summary(t Task) string {
	fields := map[string]string{}
	for _, f := range t.Fields {
		fields[f.Name] = f.Value
	}
	line := trim(expand(fields["command"]), 80)
	var notes []string
	if fields["cache"] == "true" {
		notes = append(notes, "cached")
	}
	if dep := fields["dependsOn"]; dep != "—" && dep != "" {
		notes = append(notes, "after "+dep)
	}
	if len(notes) > 0 {
		line += "  (" + strings.Join(notes, ", ") + ")"
	}
	return line
}

// display adds to a field the meaning of what a newcomer cannot read:
// ^build, {smith}, the default inputs.
func display(f Field) string {
	switch f.Name {
	case "command":
		return expand(f.Value)
	case "dependsOn":
		parts := strings.Split(f.Value, ", ")
		for i, p := range parts {
			if strings.HasPrefix(p, "^") {
				parts[i] = p + " (" + p[1:] + " of the projects it imports, first)"
			} else if p != "—" && !strings.Contains(p, ":") {
				parts[i] = p + " (this project's " + p + ", first)"
			}
		}
		return strings.Join(parts, ", ")
	}
	return f.Value
}

// expand writes {smith} as the command people type.
func expand(command string) string {
	return strings.ReplaceAll(command, "{smith}", "smith")
}

func trim(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func project(id string) string {
	name, _, _ := strings.Cut(id, ":")
	return name
}

func taskName(id string) string {
	_, name, _ := strings.Cut(id, ":")
	return name
}
