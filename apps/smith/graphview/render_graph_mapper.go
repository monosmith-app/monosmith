package graphview

import (
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// Node is a project as the graph views show it; it is also the JSON contract
// of `smith graph --json` and `smith projects --json`. Dir, Ecosystem and
// Targets repeat Root, the package ecosystem and Tasks under the names the
// smith VS Code extension reads (it was written against the TypeScript smith).
type Node struct {
	Name      string   `json:"name"`
	Root      string   `json:"root"`
	Dir       string   `json:"dir"`
	Kind      string   `json:"kind"`
	Language  string   `json:"language"`
	Ecosystem string   `json:"ecosystem"`
	Plugin    string   `json:"plugin"`
	Stack     string   `json:"stack,omitempty"`
	Version   string   `json:"version,omitempty"`
	Deps      []string `json:"deps"`
	Tasks     []string `json:"tasks"`
	Targets   []string `json:"targets"`
	Tags      []string `json:"tags,omitempty"`
	// Marked is set on the projects a view is about: the one --focus names,
	// or the changed ones of --affected.
	Marked bool `json:"marked,omitempty"`
}

// ecosystems names a language plugin's package ecosystem as the TypeScript
// smith did; any other plugin's ecosystem is its name.
var ecosystems = map[string]string{"node": "npm", "python": "pip", "go": "go"}

// Ecosystem is a plugin's package ecosystem.
func Ecosystem(plugin string) string {
	if e, ok := ecosystems[plugin]; ok {
		return e
	}
	return plugin
}

// Nodes converts the graph, sorted by name.
func Nodes(g *projectgraph.Graph, marked map[string]bool) []Node {
	out := make([]Node, 0, len(g.Projects))
	for _, p := range g.Projects {
		tasks := make([]string, 0, len(p.Tasks))
		for t := range p.Tasks {
			tasks = append(tasks, t)
		}
		sort.Strings(tasks)
		deps := p.Deps
		if deps == nil {
			deps = []string{}
		}
		out = append(out, Node{
			Name: p.Name, Root: p.Root, Dir: p.Root, Kind: string(p.Kind), Language: p.Language, Ecosystem: Ecosystem(p.Plugin),
			Plugin: p.Plugin, Stack: p.Stack, Version: p.Version, Deps: deps, Tasks: tasks, Targets: tasks, Tags: p.Tags,
			Marked: marked[p.Name],
		})
	}
	return out
}

// JSON renders {"projects": [...]}.
func JSON(g *projectgraph.Graph, marked map[string]bool) string {
	raw, _ := json.MarshalIndent(map[string]any{"projects": Nodes(g, marked)}, "", "  ")
	return string(raw) + "\n"
}

// Text renders each project with the projects it depends on beneath it; a
// marked project is starred.
func Text(g *projectgraph.Graph, marked map[string]bool) string {
	var b strings.Builder
	if len(g.Projects) == 0 {
		return "No projects yet. Add one: smith add app web --lang node\n"
	}
	for _, n := range Nodes(g, marked) {
		star := ""
		if n.Marked {
			star = " *"
		}
		fmt.Fprintf(&b, "%s%s  %s · %s · %s\n", n.Name, star, n.Kind, n.Language, n.Root)
		for i, d := range n.Deps {
			branch := "├─"
			if i == len(n.Deps)-1 {
				branch = "└─"
			}
			fmt.Fprintf(&b, "  %s %s\n", branch, d)
		}
	}
	return b.String()
}

// DOT renders Graphviz; a marked project is drawn bold.
func DOT(g *projectgraph.Graph, marked map[string]bool) string {
	var b strings.Builder
	b.WriteString("digraph smith {\n  rankdir=BT;\n  node [shape=box, style=rounded];\n")
	for _, n := range Nodes(g, marked) {
		bold := ""
		if n.Marked {
			bold = `, style="rounded,bold", penwidth=2`
		}
		fmt.Fprintf(&b, "  %q [label=%q%s];\n", n.Name, n.Name+"\n"+n.Kind+" · "+n.Language, bold)
		for _, d := range n.Deps {
			fmt.Fprintf(&b, "  %q -> %q;\n", n.Name, d)
		}
	}
	b.WriteString("}\n")
	return b.String()
}

// SVG renders the graph as an inline SVG: projects in layers, each above the
// projects it depends on, with arrows to them; a marked project is filled.
// Each box says the project's name, kind and language in text, so the colour
// is never the only signal. The SVG has a title and a description.
func SVG(g *projectgraph.Graph, marked map[string]bool) string {
	nodes := Nodes(g, marked)
	layer := map[string]int{}
	var depth func(string) int
	depth = func(name string) int {
		if l, ok := layer[name]; ok {
			return l
		}
		layer[name] = 0
		d := 0
		for _, dep := range g.Get(name).Deps {
			d = max(d, depth(dep)+1)
		}
		layer[name] = d
		return d
	}
	maxLayer := 0
	for _, n := range nodes {
		maxLayer = max(maxLayer, depth(n.Name))
	}
	const w, h, gapX, gapY = 170, 52, 30, 70
	rows := make([][]Node, maxLayer+1)
	for _, n := range nodes {
		rows[maxLayer-layer[n.Name]] = append(rows[maxLayer-layer[n.Name]], n)
	}
	widest := 1
	for _, r := range rows {
		widest = max(widest, len(r))
	}
	width, height := widest*(w+gapX)+gapX, len(rows)*(h+gapY)+gapY
	pos := map[string][2]int{}
	for y, r := range rows {
		offset := (width - len(r)*(w+gapX) + gapX) / 2
		for x, n := range r {
			pos[n.Name] = [2]int{offset + x*(w+gapX), gapY + y*(h+gapY)}
		}
	}
	colors := map[string]string{"app": "#2563eb", "lib": "#64748b", "package": "#16a34a"}
	var svg strings.Builder
	fmt.Fprintf(&svg, `<svg xmlns="http://www.w3.org/2000/svg" role="img" aria-labelledby="graph-title graph-desc" viewBox="0 0 %d %d" width="%d" height="%d"><title id="graph-title">Project dependency graph</title><desc id="graph-desc">Each box is a project; an arrow points from a project to one it depends on.</desc><defs><marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto"><path d="M0,0 L10,5 L0,10 z" fill="currentColor"/></marker></defs>`, width, height, width, height)
	for _, n := range nodes {
		from := pos[n.Name]
		for _, d := range n.Deps {
			to := pos[d]
			fmt.Fprintf(&svg, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="currentColor" stroke-opacity=".45" marker-end="url(#arrow)"/>`, from[0]+w/2, from[1]+h, to[0]+w/2, to[1])
		}
	}
	for _, n := range nodes {
		p := pos[n.Name]
		fill := "none"
		if n.Marked {
			fill = colors[n.Kind] + "33"
		}
		fmt.Fprintf(&svg, `<g><title>%s</title><rect x="%d" y="%d" width="%d" height="%d" rx="8" fill="%s" stroke="%s" stroke-width="2"/><text x="%d" y="%d" text-anchor="middle" font-weight="600">%s</text><text x="%d" y="%d" text-anchor="middle" font-size="12" opacity=".7">%s · %s</text></g>`,
			html.EscapeString(n.Root), p[0], p[1], w, h, fill, colors[n.Kind], p[0]+w/2, p[1]+22, html.EscapeString(n.Name), p[0]+w/2, p[1]+40, n.Kind, html.EscapeString(n.Language))
	}
	svg.WriteString("</svg>")
	return svg.String()
}

// HTML renders a self-contained page around the SVG of the graph.
func HTML(g *projectgraph.Graph, title string, marked map[string]bool) string {
	nodes := Nodes(g, marked)
	svg := SVG(g, marked)
	return fmt.Sprintf(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>%[1]s project graph</title>
<style>
:root { color-scheme: light dark; --bg: #ffffff; --fg: #0f172a; }
@media (prefers-color-scheme: dark) { :root { --bg: #0b1020; --fg: #e2e8f0; } }
body { margin: 0; padding: 24px 16px; background: var(--bg); color: var(--fg); font: 14px/1.4 system-ui, sans-serif; }
h1 { font-size: 18px; margin: 0 0 4px; } p { margin: 0 0 16px; opacity: .7; }
.graph { overflow-x: auto; } svg { color: var(--fg); fill: var(--fg); max-width: none; }
</style></head><body>
<h1>%[1]s</h1><p>%[2]d projects · arrows point to what a project depends on · blue app, grey lib, green package</p>
<div class="graph">%[3]s</div>
</body></html>
`, html.EscapeString(title), len(nodes), svg)
}
