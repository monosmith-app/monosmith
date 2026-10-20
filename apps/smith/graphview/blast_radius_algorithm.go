package graphview

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// Reached is a project a change reaches: the project it came through and
// whether that edge is an import or a shared schema.
type Reached struct {
	Name string `json:"name"`
	// From is the project this one depends on that the change passed through.
	From string `json:"from"`
	// Via is "import" or "schema".
	Via string `json:"via"`
}

// Radius is everything a change to a project or a file reaches.
type Radius struct {
	Target string `json:"target"`
	// Roots are the projects that own the target: where the change starts.
	Roots []string `json:"roots"`
	// Reached are the projects that depend on a root, transitively, across
	// languages, nearest first.
	Reached []Reached `json:"reached"`
}

// BlastRadius lists everything a change to target reaches. The target is a
// project name or a workspace-relative file: a file starts at the project that
// holds it and at the project that owns it as a schema.
func BlastRadius(g *projectgraph.Graph, target string) (Radius, error) {
	target = strings.TrimPrefix(strings.ReplaceAll(target, "\\", "/"), "./")
	r := Radius{Target: target}
	if p := g.Get(target); p != nil {
		r.Roots = []string{p.Name}
	} else {
		if p := g.Owner(target); p != nil {
			r.Roots = append(r.Roots, p.Name)
		}
		if p := g.SchemaOwner(target); p != nil && !slices.Contains(r.Roots, p.Name) {
			r.Roots = append(r.Roots, p.Name)
		}
	}
	if len(r.Roots) == 0 {
		return r, fmt.Errorf("%q is neither a project nor a file inside one", target)
	}
	sort.Strings(r.Roots)
	seen := map[string]bool{}
	for _, root := range r.Roots {
		seen[root] = true
	}
	queue := append([]string(nil), r.Roots...)
	for len(queue) > 0 {
		from := queue[0]
		queue = queue[1:]
		dependents := g.Dependents(from)
		sort.Strings(dependents)
		for _, name := range dependents {
			if seen[name] {
				continue
			}
			seen[name] = true
			via := "import"
			if slices.Contains(g.Get(name).SchemaDeps, from) {
				via = "schema"
			}
			r.Reached = append(r.Reached, Reached{Name: name, From: from, Via: via})
			queue = append(queue, name)
		}
	}
	return r, nil
}

// BlastRadiusText prints a radius, one project per line.
func BlastRadiusText(r Radius) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s starts at %s\n", r.Target, strings.Join(r.Roots, ", "))
	if len(r.Reached) == 0 {
		b.WriteString("nothing else depends on it\n")
		return b.String()
	}
	for _, x := range r.Reached {
		fmt.Fprintf(&b, "  %-20s via %s of %s\n", x.Name, x.Via, x.From)
	}
	return b.String()
}
