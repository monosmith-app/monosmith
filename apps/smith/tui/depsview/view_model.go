package depsview

import (
	"slices"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
)

// Model lists the dependencies that have newer releases, patch before
// minor before major, and the one-version policy's verdict. Space marks a row;
// updating ends as `smith deps up <packages>`, aligning as `smith deps up --align`.
type Model struct {
	Loaded bool
	// Waited is how long the load has taken so far, set by the shell for drawing.
	Waited time.Duration
	// Projects is how many projects the load covers, set by the shell for drawing (0 when unknown).
	Projects int
	Rows     []Row
	Policy   []PolicyLine
	Err      error
	Cursor   int
	Marked   map[string]bool
	// Anchor is the row a shift-click or shift-arrow extends the marks from.
	Anchor int
	// scroll is where the list's window starts, apart from the cursor, so the
	// wheel can scroll without moving it (see pane.Scroll).
	scroll *pane.Scroll
}

// HeadingRow is the row number the table's heading stands for in the map View
// returns: its box marks or unmarks every row.
const HeadingRow = -1

// BoxColumns is where the box of a row is, in cells from the left of the view.
const (
	BoxFrom = 2
	BoxTo   = 5
)

// FromResult is the model for a loaded result.
func FromResult(r Result) Model {
	// Patch and minor first, majors last, whatever order they were found in.
	rows := slices.Clone(r.Rows)
	slices.SortStableFunc(rows, func(a, b Row) int {
		return levelRank(a.Level) - levelRank(b.Level)
	})
	return Model{Loaded: true, Rows: rows, Policy: r.Policy, Err: r.Err, scroll: &pane.Scroll{}}
}

func levelRank(level string) int {
	switch level {
	case "patch":
		return 0
	case "minor":
		return 1
	}
	return 2
}

// Update handles a key; page is how many rows a page key moves.
func (v Model) Update(key tea.KeyMsg, page int) Model {
	switch key.String() {
	case " ", "space":
		return v.toggle()
	case "a":
		return v.markAll()
	case "s":
		return v.markSafe()
	case "shift+up":
		return v.extend(-1)
	case "shift+down":
		return v.extend(1)
	}
	v.Cursor = pane.MoveCursor(v.Cursor, len(v.Rows), commands.Nav(key.String()), page)
	v.Anchor = v.Cursor
	return v
}

// Wheel scrolls the list by delta rows; the cursor and the marks stay.
func (v Model) Wheel(delta int) Model {
	if v.scroll == nil {
		v.scroll = &pane.Scroll{}
	}
	v.scroll.By(delta)
	return v
}

// extend moves the cursor by delta and marks exactly the rows from the anchor
// to it, so reversing the arrow shrinks the marks again.
func (v Model) extend(delta int) Model {
	if len(v.Rows) == 0 {
		return v
	}
	v.Anchor = min(max(v.Anchor, 0), len(v.Rows)-1)
	v.Cursor = pane.MoveCursor(v.Cursor, len(v.Rows), map[int]commands.NavAction{-1: commands.NavUp, 1: commands.NavDown}[delta], 1)
	return v.markRange(v.Anchor, v.Cursor, true, false)
}

// markRange marks (or, with on false, unmarks) the rows from a to b, both
// included; with replace the marks elsewhere are dropped first.
func (v Model) markRange(a, b int, on, keep bool) Model {
	if a > b {
		a, b = b, a
	}
	marked := map[string]bool{}
	if keep {
		for k := range v.Marked {
			marked[k] = true
		}
	}
	for i := max(a, 0); i <= b && i < len(v.Rows); i++ {
		if on {
			marked[v.Rows[i].Name] = true
		} else {
			delete(marked, v.Rows[i].Name)
		}
	}
	v.Marked = marked
	return v
}

// MarkState is 0 when no row is marked, 2 when every row is, and 1 between.
func (v Model) MarkState() int {
	n := 0
	for _, r := range v.Rows {
		if v.Marked[r.Name] {
			n++
		}
	}
	switch {
	case n == 0:
		return 0
	case n == len(v.Rows):
		return 2
	}
	return 1
}

// TickRow is a click on the box of row i: it flips the row's mark, or, with
// extend, marks or unmarks every row from the anchor to it as that row will be.
func (v Model) TickRow(i int, extend bool) Model {
	if i < 0 || i >= len(v.Rows) {
		return v
	}
	on := !v.Marked[v.Rows[i].Name]
	if extend {
		v = v.markRange(v.Anchor, i, on, true)
	} else {
		v.Cursor = i
		v.Anchor = i
		v.Marked = withMark(v.Marked, v.Rows[i].Name, on)
		return v
	}
	v.Cursor = i
	return v
}

// ChooseRow is a click on a row away from its box: the cursor goes there and
// nothing is marked or unmarked, or, with extend, the rows from the anchor to
// it are marked (the anchor stays, as in a list of mail).
func (v Model) ChooseRow(i int, extend bool) Model {
	if i < 0 || i >= len(v.Rows) {
		return v
	}
	if extend {
		v = v.markRange(v.Anchor, i, true, true)
	} else {
		v.Anchor = i
	}
	v.Cursor = i
	return v
}

// TickAll is a click on the box of the heading: it marks every row, or
// unmarks them all when every row is marked.
func (v Model) TickAll() Model { return v.markAll() }

func withMark(marked map[string]bool, name string, on bool) map[string]bool {
	out := map[string]bool{}
	for k := range marked {
		out[k] = true
	}
	if on {
		out[name] = true
	} else {
		delete(out, name)
	}
	return out
}

// toggle marks or unmarks the row under the cursor.
func (v Model) toggle() Model {
	if v.Cursor < 0 || v.Cursor >= len(v.Rows) {
		return v
	}
	marked := map[string]bool{}
	for k := range v.Marked {
		marked[k] = true
	}
	name := v.Rows[v.Cursor].Name
	if marked[name] {
		delete(marked, name)
	} else {
		marked[name] = true
	}
	v.Marked = marked
	return v
}

// markSafe marks every patch and minor row (what can be taken together), or
// unmarks them when they are all marked already. Marked majors stay as they are.
func (v Model) markSafe() Model {
	marked := map[string]bool{}
	for k := range v.Marked {
		marked[k] = true
	}
	all := true
	for _, r := range v.Rows {
		if r.Level != "major" && !marked[r.Name] {
			all = false
		}
	}
	for _, r := range v.Rows {
		if r.Level == "major" {
			continue
		}
		if all {
			delete(marked, r.Name)
		} else {
			marked[r.Name] = true
		}
	}
	v.Marked = marked
	return v
}

// chosenMajors counts the major releases among the packages an update would move.
func (v Model) chosenMajors() int {
	chosen := map[string]bool{}
	for _, n := range v.chosen() {
		chosen[n] = true
	}
	n := 0
	for _, r := range v.Rows {
		if r.Level == "major" && chosen[r.Name] {
			n++
		}
	}
	return n
}

// markAll marks every row, or none when every row is already marked.
func (v Model) markAll() Model {
	marked := map[string]bool{}
	if len(v.Marked) < len(v.Rows) {
		for _, r := range v.Rows {
			marked[r.Name] = true
		}
	}
	v.Marked = marked
	return v
}

// chosen are the packages an update would move: the marked rows, or the row
// under the cursor when none is marked.
func (v Model) chosen() []string {
	var names []string
	for _, r := range v.Rows {
		if v.Marked[r.Name] {
			names = append(names, r.Name)
		}
	}
	if len(names) == 0 && v.Cursor >= 0 && v.Cursor < len(v.Rows) {
		names = append(names, v.Rows[v.Cursor].Name)
	}
	return names
}

// UpArgv is the command line that updates the chosen packages; nil when none.
func (v Model) UpArgv() []string {
	names := v.chosen()
	if len(names) == 0 {
		return nil
	}
	argv := append([]string{"deps", "up"}, names...)
	for _, r := range v.Rows {
		if r.Level == "major" && slices.Contains(names, r.Name) {
			// A major update was marked on purpose: the command takes it only with --major.
			return append(argv, "--major")
		}
	}
	return argv
}

// AlignArgv is the command line that brings the policy's drifting dependencies
// to one version; nil unless the policy found some.
func (v Model) AlignArgv() []string {
	if v.drifting() {
		return []string{"deps", "up", "--align"}
	}
	return nil
}

// drifting is true when the one-version policy found dependencies at different versions.
func (v Model) drifting() bool {
	return slices.ContainsFunc(v.Policy, func(p PolicyLine) bool { return !p.OK && !p.Warning })
}
