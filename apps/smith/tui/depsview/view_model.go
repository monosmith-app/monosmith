package depsview

import (
	"fmt"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Model lists the dependencies that have newer releases, patch before
// minor before major, and the one-version policy's verdict. Space marks a row;
// updating ends as `smith deps up <packages>`, aligning as `smith deps up --align`.
type Model struct {
	Loaded bool
	// Waited is how long the load has taken so far, set by the shell for drawing.
	Waited time.Duration
	Rows   []Row
	Policy []PolicyLine
	Err    error
	Cursor int
	Marked map[string]bool
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
	v.Cursor = pane.MoveCursor(v.Cursor, len(v.Rows), key.String(), page)
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
	v.Cursor = pane.MoveCursor(v.Cursor, len(v.Rows), map[int]string{-1: "up", 1: "down"}[delta], 1)
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
	return append([]string{"deps", "up"}, names...)
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

// levelCounts counts the rows of each level.
func (v Model) levelCounts() (patch, minor, major int) {
	for _, r := range v.Rows {
		switch r.Level {
		case "patch":
			patch++
		case "minor":
			minor++
		default:
			major++
		}
	}
	return
}

// View draws the view in a w by h area; rowOf maps a line to the row drawn on it.
func (v Model) View(th theme.Theme, w, h int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	switch {
	case v.Err != nil:
		return []string{th.Failure(v.Err.Error())}, rowOf
	case !v.Loaded:
		return pane.Wrap(th.Muted(progress.Waiting("Asking the package registries for newer releases of every dependency.", v.Waited)+" On a large workspace this takes up to half a minute; you can use the other tabs meanwhile."), w, 0), rowOf
	}
	var tail []string
	tail = append(tail, "", th.Title("One version"))
	if len(v.Policy) == 0 {
		tail = append(tail, "  "+th.Muted("not required yet: press o to make every project use the same version of a shared dependency"))
	}
	for _, p := range v.Policy {
		mark := th.Success(th.Icons.Passed)
		switch {
		case !p.OK && p.Warning:
			mark = th.Warning(th.Icons.Warning)
		case !p.OK:
			mark = th.Failure(th.Icons.Failed)
		}
		tail = append(tail, "  "+mark+" "+p.Check)
		if !p.OK && p.Detail != "" {
			for _, d := range strings.Split(p.Detail, "; ") {
				tail = append(tail, "      "+th.Muted(d))
			}
		}
	}
	if n := v.chosenMajors(); n > 0 {
		tail = append(tail, "", th.Warning(th.Icons.Warning+" "+fmt.Sprintf("%d of the chosen packages are major releases: they can break your code. Read their release notes first.", n)))
	}
	var wrapped []string
	for _, l := range tail {
		wrapped = append(wrapped, pane.Wrap(l, w, 4)...)
	}
	tail = wrapped

	if len(v.Rows) == 0 {
		lines = append(lines, th.Success(th.Icons.Passed+" Every dependency is on its latest release."))
	} else {
		patch, minor, major := v.levelCounts()
		lines = append(lines, th.Bold(fmt.Sprintf("%d with newer releases: %d patch · %d minor · %d major", len(v.Rows), patch, minor, major))+th.Muted(fmt.Sprintf("   %d marked", len(v.Marked))))
		if safe := patch + minor; safe > 0 {
			lines = append(lines, pane.Wrap(th.Success(fmt.Sprintf("%d patch and minor releases are safe to take together (press s).", safe))+th.Muted(" The "+fmt.Sprint(major)+" major ones can break your code: take them one at a time."), w, 0)...)
		}
		lines = append(lines, "")
		nameW := 8
		for _, r := range v.Rows {
			nameW = max(nameW, pane.Width(r.Name))
		}
		nameW = min(nameW+1, max(w/3, 8))
		const verW, levelW = 11, 8
		usedW := max(w-6-nameW-2*verW-levelW, 8)
		cols := func(name, now, latest, level, projects string) string {
			return pane.MiddleCell(name, nameW) + pane.Cell(now, verW) + pane.Cell(latest, verW) + pane.Cell(level, levelW) + pane.Cell(projects, usedW)
		}
		headBox := [3]string{"[ ]", "[-]", "[x]"}[v.MarkState()]
		rowOf[len(lines)] = HeadingRow
		lines = append(lines, th.Muted(pane.Fit("  "+headBox+" "+cols("Package", "Now", "Latest", "Level", "Projects"), w)))
		// Two group headings may sit among the rows: they take room too.
		avail := max(h-len(lines)-len(tail)-2, 3)
		top := v.scroll.Show(v.Cursor, avail, len(v.Rows))
		first := len(lines)
		prevMajor := top > 0 && v.Rows[top-1].Level == "major"
		for i := top; i < len(v.Rows) && i < top+avail; i++ {
			r := v.Rows[i]
			if isMajor := r.Level == "major"; i == top || isMajor != prevMajor {
				if isMajor {
					lines = append(lines, "  "+th.Warning("Major releases: may break your code, take one at a time"))
				} else {
					lines = append(lines, "  "+th.Success("Patch and minor releases: safe to take together"))
				}
			}
			prevMajor = r.Level == "major"
			box := "[ ] "
			if v.Marked[r.Name] {
				box = "[x] "
			}
			level := th.Muted(r.Level)
			switch r.Level {
			case "major":
				level = th.Warning(r.Level)
			case "patch":
				level = th.Success(r.Level)
			}
			projects := strings.Join(r.Projects, ", ")
			plain := "  " + box + cols(r.Name, r.Current, r.Latest, r.Level, projects)
			line := "  " + box + cols(r.Name, r.Current, r.Latest, level, projects)
			if i == v.Cursor {
				line = th.Selected(pane.Fit(plain, w))
			}
			rowOf[len(lines)] = i
			lines = append(lines, line)
		}
		if len(v.Rows) > avail {
			// The rows and the group headings between them scroll as one block.
			block := pane.Scrollbar(th, lines[first:], w, top, len(v.Rows)+len(lines[first:])-min(avail, len(v.Rows)-top))
			copy(lines[first:], block)
		}
	}
	lines = append(lines, tail...)
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, rowOf
}
