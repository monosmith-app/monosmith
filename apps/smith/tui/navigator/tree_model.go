package navigator

import (
	"sort"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// groupOrder is the order the kinds are listed in; other kinds follow.
var groupOrder = []string{"app", "lib", "package"}

// Tree is the navigator. It is a value: every method returns the new tree.
type Tree struct {
	projects []Project
	byName   map[string]Project
	// open holds the expanded keys: "group:<kind>" and "project:<name>".
	open   map[string]bool
	cursor int
	top    int
	// height is the number of rows that fit.
	height int
}

// New builds the tree; every group starts open and every project closed.
func New(projects []Project) Tree {
	t := Tree{projects: projects, byName: map[string]Project{}, open: map[string]bool{}, height: 10}
	for _, p := range projects {
		t.byName[p.Name] = p
		t.open["group:"+p.Kind] = true
	}
	return t
}

// Projects returns the projects.
func (t Tree) Projects() []Project { return t.projects }

// Project returns the project with a name.
func (t Tree) Project(name string) (Project, bool) { p, ok := t.byName[name]; return p, ok }

// SetHeight tells the tree how many rows fit; when that changes it keeps the
// cursor visible. Setting the same height leaves the scroll position alone, so
// drawing never undoes a wheel scroll.
func (t Tree) SetHeight(h int) Tree {
	h = max(h, 1)
	if h == t.height {
		return t
	}
	t.height = h
	return t.reveal()
}

// Rows are the visible rows.
func (t Tree) Rows() []Row {
	groups := map[string][]Project{}
	for _, p := range t.projects {
		groups[p.Kind] = append(groups[p.Kind], p)
	}
	var kinds []string
	for _, k := range groupOrder {
		if len(groups[k]) > 0 {
			kinds = append(kinds, k)
		}
	}
	var others []string
	for k := range groups {
		if !contains(groupOrder, k) {
			others = append(others, k)
		}
	}
	sort.Strings(others)
	kinds = append(kinds, others...)

	var rows []Row
	for _, k := range kinds {
		open := t.open["group:"+k]
		rows = append(rows, Row{Kind: GroupRow, Group: k, Expandable: true, Expanded: open})
		if !open {
			continue
		}
		ps := groups[k]
		sort.Slice(ps, func(i, j int) bool { return ps[i].Name < ps[j].Name })
		for _, p := range ps {
			popen := t.open["project:"+p.Name]
			rows = append(rows, Row{Kind: ProjectRow, Project: p.Name, Depth: 1, Expandable: len(p.Tasks) > 0, Expanded: popen})
			if popen {
				for _, task := range p.Tasks {
					rows = append(rows, Row{Kind: TaskRow, Project: p.Name, Task: task.Name, Depth: 2})
				}
			}
		}
	}
	return rows
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// Cursor is the index of the selected row.
func (t Tree) Cursor() int { return t.cursor }

// Selected is the row under the cursor.
func (t Tree) Selected() Selection {
	rows := t.Rows()
	if t.cursor < 0 || t.cursor >= len(rows) {
		return Selection{}
	}
	r := rows[t.cursor]
	return Selection{Kind: r.Kind, Group: r.Group, Project: r.Project, Task: r.Task, Valid: true}
}

// SelectProject moves the cursor to a project (opening its group).
func (t Tree) SelectProject(name string) Tree {
	if p, ok := t.byName[name]; ok {
		t.open = cloneSet(t.open)
		t.open["group:"+p.Kind] = true
	}
	for i, r := range t.Rows() {
		if r.Kind == ProjectRow && r.Project == name {
			t.cursor = i
		}
	}
	return t.reveal()
}

// Move moves the cursor by delta rows, clamped.
func (t Tree) Move(delta int) Tree {
	t.cursor = min(max(t.cursor+delta, 0), max(len(t.Rows())-1, 0))
	return t.reveal()
}

// Home and End jump to the first and last row.
func (t Tree) Home() Tree { return t.Move(-len(t.Rows())) }
func (t Tree) End() Tree  { return t.Move(len(t.Rows())) }

// Page moves by a page.
func (t Tree) Page(dir int) Tree { return t.Move(dir * max(t.height-1, 1)) }

// Scroll moves the viewport without moving the cursor (the mouse wheel).
func (t Tree) Scroll(delta int) Tree {
	t.top = min(max(t.top+delta, 0), max(len(t.Rows())-t.height, 0))
	return t
}

// Select moves the cursor to the row at index i of the visible rows.
func (t Tree) Select(i int) Tree {
	if i >= 0 && i < len(t.Rows()) {
		t.cursor = i
	}
	return t.reveal()
}

// RowAt is the index of the row drawn at line y of the tree's area (0 is the
// first line), or -1.
func (t Tree) RowAt(y int) int {
	i := t.top + y
	if y < 0 || y >= t.height || i >= len(t.Rows()) {
		return -1
	}
	return i
}

// Toggle opens or closes the group or project under the cursor.
func (t Tree) Toggle() Tree {
	rows := t.Rows()
	if t.cursor >= len(rows) {
		return t
	}
	return t.setOpen(rows[t.cursor], !rows[t.cursor].Expanded)
}

// Expand opens the row under the cursor; Collapse closes it, or moves to its parent.
func (t Tree) Expand() Tree {
	rows := t.Rows()
	if t.cursor < len(rows) && rows[t.cursor].Expandable && !rows[t.cursor].Expanded {
		return t.setOpen(rows[t.cursor], true)
	}
	return t
}

// Collapse closes the row under the cursor, or moves up to its parent when it is closed.
func (t Tree) Collapse() Tree {
	rows := t.Rows()
	if t.cursor >= len(rows) {
		return t
	}
	if r := rows[t.cursor]; r.Expandable && r.Expanded {
		return t.setOpen(r, false)
	}
	for i := t.cursor - 1; i >= 0; i-- {
		if rows[i].Depth < rows[t.cursor].Depth {
			t.cursor = i
			break
		}
	}
	return t.reveal()
}

func (t Tree) setOpen(r Row, open bool) Tree {
	t.open = cloneSet(t.open)
	switch r.Kind {
	case GroupRow:
		t.open["group:"+r.Group] = open
	case ProjectRow:
		t.open["project:"+r.Project] = open
	}
	return t.reveal()
}

func cloneSet(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m)+1)
	for k, v := range m {
		out[k] = v
	}
	return out
}

// reveal scrolls so the cursor is on screen and the top is valid.
func (t Tree) reveal() Tree {
	n := len(t.Rows())
	t.cursor = min(max(t.cursor, 0), max(n-1, 0))
	if t.cursor < t.top {
		t.top = t.cursor
	}
	if t.cursor >= t.top+t.height {
		t.top = t.cursor - t.height + 1
	}
	t.top = min(max(t.top, 0), max(n-t.height, 0))
	return t
}

// View draws the visible rows into w cells and height lines. hover is the row
// index under the pointer, or -1; focused says whether the pane has the keyboard.
func (t Tree) View(th theme.Theme, w int, hover int, focused bool, status func(project, task string) string) []string {
	rows := t.Rows()
	if len(rows) == 0 {
		return []string{th.Muted(" No projects yet.")}
	}
	var out []string
	for i := t.top; i < len(rows) && len(out) < t.height; i++ {
		out = append(out, t.line(th, rows[i], w, i == t.cursor, i == hover, focused, status))
	}
	return out
}

func (t Tree) line(th theme.Theme, r Row, w int, selected, hovered, focused bool, status func(project, task string) string) string {
	indent := strings.Repeat("  ", r.Depth)
	arrow := "  "
	if r.Expandable {
		arrow = th.Icons.Closed + " "
		if r.Expanded {
			arrow = th.Icons.Open + " "
		}
	}
	var text string
	switch r.Kind {
	case GroupRow:
		text = indent + arrow + strings.ToUpper(r.Group[:min(1, len(r.Group))]) + r.Group[min(1, len(r.Group)):] + "s"
	case ProjectRow:
		p := t.byName[r.Project]
		dot := th.Muted(th.Icons.Queued)
		if p.Affected {
			dot = th.Warning(th.Icons.Dot)
		}
		suffix := ""
		if status != nil {
			suffix = status(r.Project, "")
		}
		head := indent + arrow + dot + " "
		// The name matters more than the language, and neither may run off
		// the pane: the language goes first, then the name ends in an ellipsis.
		room := w - 1 - pane.Width(head) - pane.Width(suffix)
		name, lang := r.Project, ""
		if p.Language != "" && pane.Width(name)+1+pane.Width(p.Language) <= room {
			lang = " " + th.Muted(p.Language)
		}
		if pane.Width(name) > room && room > 1 {
			name = ansi.Truncate(name, room, "…")
		}
		text = head + name + lang + suffix
	case TaskRow:
		suffix := ""
		if status != nil {
			suffix = status(r.Project, r.Task)
		}
		text = indent + arrow + th.Icons.Pointer + " " + r.Task + suffix
	}
	switch {
	case selected && focused:
		return th.Selected(pane.Fit(" "+ansi.Strip(text), w))
	case selected, hovered:
		return th.Hover(pane.Fit(" "+ansi.Strip(text), w))
	}
	return pane.Fit(" "+text, w)
}

// Ordered returns the projects in the order the tree lists them: by kind
// (apps, libs, packages, then others), then by name.
func (t Tree) Ordered() []Project {
	var out []Project
	seen := map[string]bool{}
	for _, r := range t.openAll().Rows() {
		if r.Kind == ProjectRow && !seen[r.Project] {
			seen[r.Project] = true
			out = append(out, t.byName[r.Project])
		}
	}
	return out
}

// openAll is the tree with every group open, to list all projects.
func (t Tree) openAll() Tree {
	t.open = cloneSet(t.open)
	for _, p := range t.projects {
		t.open["group:"+p.Kind] = true
	}
	return t
}

// CursorLine is the line, from the top of the visible rows, the cursor is on.
func (t Tree) CursorLine() int { return min(max(t.cursor-t.top, 0), t.height-1) }
