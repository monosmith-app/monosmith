package cipipeline

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// TrackedChange is a tracked file left modified or deleted after verify.
type TrackedChange struct {
	File string
	// Project is the project whose directory holds the file, "" for one
	// outside every project (the workspace root).
	Project string
	// Tasks are the project's tasks verify ran, the likely writers.
	Tasks []string
	// Before is true when the file was already changed before verify
	// started: setup (an install) changed it, not a verify task.
	Before bool
}

// TrackedChanges are the tracked files changed after verify, from the
// working tree's changed files before and after it (release.Uncommitted of
// `git status --porcelain --untracked-files=no`). A release refuses a dirty
// tree, so each one would stop the release of this commit. ran lists the
// tasks verify ran, as project:task.
func TrackedChanges(before, after []string, projects []*projectgraph.Project, ran []string) []TrackedChange {
	tasksOf := map[string][]string{}
	for _, id := range ran {
		project, task, ok := strings.Cut(id, ":")
		if ok && !slices.Contains(tasksOf[project], task) {
			tasksOf[project] = append(tasksOf[project], task)
		}
	}
	var out []TrackedChange
	for _, file := range after {
		project := owner(file, projects)
		tasks := slices.Clone(tasksOf[project])
		sort.Strings(tasks)
		out = append(out, TrackedChange{File: file, Project: project, Tasks: tasks, Before: slices.Contains(before, file)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].File < out[j].File })
	return out
}

// owner is the project whose root holds file, the deepest one when roots
// nest.
func owner(file string, projects []*projectgraph.Project) string {
	best, depth := "", -1
	for _, p := range projects {
		root := strings.Trim(p.Root, "/")
		if root == "" || root == "." {
			continue
		}
		if (file == root || strings.HasPrefix(file, root+"/")) && len(root) > depth {
			best, depth = p.Name, len(root)
		}
	}
	return best
}

// ExplainTrackedChanges says which tracked files verify left changed, who
// probably changed them, and what to do: the error a pull request fails
// with, before the same tree stops the release on the release branch.
func ExplainTrackedChanges(changes []TrackedChange) string {
	var b strings.Builder
	b.WriteString("tracked files are changed after verify, so a release from this commit would refuse the dirty working tree:\n")
	for _, c := range changes {
		who := "verify's tasks"
		switch {
		case c.Before:
			who = "already changed before verify started (by setup, e.g. an install rewriting a lockfile)"
		case c.Project != "" && len(c.Tasks) > 0:
			who = fmt.Sprintf("in %s, whose %s ran", c.Project, strings.Join(c.Tasks, ", "))
		case c.Project != "":
			who = "in " + c.Project
		}
		fmt.Fprintf(&b, "  - %s (%s)\n", c.File, who)
	}
	b.WriteString("A file a task generates (build state such as *.tsbuildinfo, a cache, an output) belongs out of git: add it to .gitignore and `git rm --cached` it. A file a task updates on purpose (a lockfile, generated source) is committed with the change that needs it.")
	return b.String()
}
