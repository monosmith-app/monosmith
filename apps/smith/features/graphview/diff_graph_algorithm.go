package graphview

import (
	"fmt"
	"sort"
	"strings"
)

// GraphDiff is how the project graph changed between two revisions.
type GraphDiff struct {
	From            string      `json:"from"`
	To              string      `json:"to"`
	AddedProjects   []Node      `json:"addedProjects"`
	RemovedProjects []Node      `json:"removedProjects"`
	AddedEdges      [][2]string `json:"addedEdges"`
	RemovedEdges    [][2]string `json:"removedEdges"`
	Changed         []Change    `json:"changed"`
}

// Change is one property of a project that differs.
type Change struct {
	Project string `json:"project"`
	Field   string `json:"field"`
	Before  string `json:"before"`
	After   string `json:"after"`
}

// Empty reports whether nothing changed.
func (d GraphDiff) Empty() bool {
	return len(d.AddedProjects)+len(d.RemovedProjects)+len(d.AddedEdges)+len(d.RemovedEdges)+len(d.Changed) == 0
}

// Diff compares two graphs' nodes: projects added and removed, edges (a
// project importing another) added and removed, and the kind, stack, plugin
// and directory of the projects in both.
func Diff(before, after []Node) GraphDiff {
	b, a := index(before), index(after)
	var d GraphDiff
	for _, n := range after {
		if _, ok := b[n.Name]; !ok {
			d.AddedProjects = append(d.AddedProjects, n)
		}
	}
	for _, n := range before {
		if _, ok := a[n.Name]; !ok {
			d.RemovedProjects = append(d.RemovedProjects, n)
		}
	}
	edges := func(nodes []Node) map[[2]string]bool {
		out := map[[2]string]bool{}
		for _, n := range nodes {
			for _, dep := range n.Deps {
				out[[2]string{n.Name, dep}] = true
			}
		}
		return out
	}
	be, ae := edges(before), edges(after)
	for e := range ae {
		if !be[e] {
			d.AddedEdges = append(d.AddedEdges, e)
		}
	}
	for e := range be {
		if !ae[e] {
			d.RemovedEdges = append(d.RemovedEdges, e)
		}
	}
	sortEdges(d.AddedEdges)
	sortEdges(d.RemovedEdges)
	for _, n := range after {
		old, ok := b[n.Name]
		if !ok {
			continue
		}
		for _, f := range []struct{ name, before, after string }{
			{"kind", old.Kind, n.Kind}, {"stack", old.Stack, n.Stack}, {"plugin", old.Plugin, n.Plugin}, {"root", old.Root, n.Root},
		} {
			if f.before != f.after {
				d.Changed = append(d.Changed, Change{Project: n.Name, Field: f.name, Before: f.before, After: f.after})
			}
		}
	}
	return d
}

func index(nodes []Node) map[string]Node {
	out := map[string]Node{}
	for _, n := range nodes {
		out[n.Name] = n
	}
	return out
}

func sortEdges(e [][2]string) {
	sort.Slice(e, func(i, j int) bool { return e[i][0]+" "+e[i][1] < e[j][0]+" "+e[j][1] })
}

// DiffText renders a diff for people: + added, - removed, ~ changed.
func DiffText(d GraphDiff) string {
	var b strings.Builder
	fmt.Fprintf(&b, "project graph from %s to %s\n", d.From, d.To)
	if d.Empty() {
		b.WriteString("  no change\n")
		return b.String()
	}
	for _, n := range d.AddedProjects {
		fmt.Fprintf(&b, "+ %s (%s %s, %s)\n", n.Name, n.Language, n.Kind, n.Root)
	}
	for _, n := range d.RemovedProjects {
		fmt.Fprintf(&b, "- %s (%s %s, %s)\n", n.Name, n.Language, n.Kind, n.Root)
	}
	for _, e := range d.AddedEdges {
		fmt.Fprintf(&b, "+ %s → %s\n", e[0], e[1])
	}
	for _, e := range d.RemovedEdges {
		fmt.Fprintf(&b, "- %s → %s\n", e[0], e[1])
	}
	for _, c := range d.Changed {
		fmt.Fprintf(&b, "~ %s %s: %s → %s\n", c.Project, c.Field, orNone(c.Before), orNone(c.After))
	}
	return b.String()
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}
