package projectpage

import (
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// kindNouns are the catalog keys of the plural nouns of the common kinds.
var kindNouns = map[string]string{
	"app":     "projectpage.kind.app",
	"lib":     "projectpage.kind.lib",
	"package": "projectpage.kind.package",
}

// kindHeadings are the catalog keys of the group headings of the common kinds.
var kindHeadings = map[string]string{
	"app":     "projectpage.heading.app",
	"lib":     "projectpage.heading.lib",
	"package": "projectpage.heading.package",
}

// Pointer is the mark of the selected row, in the accent.
const Pointer = "▶ "

// ListHead is the top of the list: the filter line, with the cursor while the
// person types, and the muted count line ("1 of 2 apps · 1 lib · 1 changed").
//
// Parameters: th is the theme; in is what the list shows; w is the width.
// Returns: two lines of w cells.
func ListHead(th theme.Theme, in ListInput, w int) []string {
	filter := th.Muted(words.Say("projectpage.filter.label")) + " "
	switch {
	case in.Typing:
		filter += in.Filter + th.Accent("▏")
	case in.Filter == "":
		filter += th.Faint(words.Say("projectpage.filter.hint"))
	default:
		filter += in.Filter
	}
	return []string{pane.Fit(pane.CutEllipsis(filter, w), w), pane.Fit(pane.CutEllipsis(th.Muted(countLine(in)), w), w)}
}

// countLine says how many projects of each kind the list shows: "N of M apps"
// while a filter is set, else "M apps"; then how many changed.
func countLine(in ListInput) string {
	shown := map[string]int{}
	for _, r := range in.Rows {
		shown[r.Kind]++
	}
	var parts []string
	for _, kind := range in.Kinds {
		total := in.Totals[kind]
		noun := words.Say("projectpage.kind.other", "count", total, "kind", kind)
		if key := kindNouns[kind]; key != "" {
			noun = words.Count(key, total)
		}
		if in.Filter != "" {
			noun = words.Say("projectpage.count.of", "shown", shown[kind], "total", noun)
		}
		parts = append(parts, noun)
	}
	if in.Changed > 0 {
		parts = append(parts, words.Say("projectpage.count.changed", "count", in.Changed))
	}
	if len(parts) == 0 {
		return words.Say("projectpage.count.none")
	}
	return strings.Join(parts, " · ")
}

// ListBody is the projects under their kind headings: a muted bold heading
// with its count for each kind, then one row per project. A row is the
// pointer of the selected row, the last-run mark, the name, the language
// (and the kind and the version when there is room) in muted text, and the
// changed mark in the accent. The selected row is on the Highlight
// background with its name in the accent and bold.
//
// Parameters: th is the theme; in is what the list shows; w is the width.
// Returns: the lines, the Index of the row on each line, and the line of the
// selected row (-1 when none is listed).
func ListBody(th theme.Theme, in ListInput, w int) (lines []string, rowOf map[int]int, selected int) {
	rowOf, selected = map[int]int{}, -1
	if len(in.Rows) == 0 {
		for _, part := range pane.Wrap(words.Say("projectpage.list.empty"), w, 0) {
			lines = append(lines, pane.Fit(th.Muted(part), w))
		}
		return lines, rowOf, -1
	}
	cols := listColumns(in.Rows, w)
	kind, count := "", map[string]int{}
	for _, r := range in.Rows {
		count[r.Kind]++
	}
	for _, r := range in.Rows {
		if r.Kind != kind || len(lines) == 0 {
			kind = r.Kind
			lines = append(lines, pane.Fit(th.Muted(th.Bold(heading(kind)+" "+strconv.Itoa(count[kind]))), w))
		}
		if r.Index == in.Selected {
			selected = len(lines)
		}
		rowOf[len(lines)] = r.Index
		lines = append(lines, listRow(th, r, cols, w, r.Index == in.Selected, in.Tick))
	}
	return lines, rowOf, selected
}

// heading is the group heading of a kind, in capitals.
func heading(kind string) string {
	if key := kindHeadings[kind]; key != "" {
		return words.Say(key)
	}
	return strings.ToUpper(kind)
}

// columns is how a row spends its cells.
type columns struct {
	name, lang, kind, version int
	// changed is the width of the changed mark: the dot and its word, or the dot alone.
	changed int
}

// leadW is the pointer, the mark and the two spaces after it.
const leadW = 5

// listColumns fits the columns into w cells: the name and the changed mark
// always, then the language, the kind and the version while they fit.
func listColumns(rows []ListRow, w int) columns {
	var c columns
	for _, r := range rows {
		c.name = max(c.name, pane.Width(r.Name)+2)
		c.lang = max(c.lang, pane.Width(r.Language)+2)
		c.kind = max(c.kind, pane.Width(r.Kind)+2)
		c.version = max(c.version, pane.Width(r.Version)+2)
	}
	c.name, c.lang = min(max(c.name, 8), 24), min(c.lang, 12)
	c.changed = pane.Width(words.Say("projectpage.changed", "dot", "●"))
	room := w - leadW - c.changed
	if room < c.name+c.lang {
		c.changed = 1
		room = w - leadW - c.changed
	}
	c.name = min(c.name, max(room-c.lang, 6))
	room -= c.name
	if room < c.lang {
		c.lang = 0
	}
	room -= c.lang
	if room < c.kind+10 {
		c.kind = 0
	}
	room -= c.kind
	if room < c.version || c.version <= 2 {
		c.version = 0
	}
	return c
}

// listRow draws one project in w cells.
func listRow(th theme.Theme, r ListRow, c columns, w int, selected bool, tick int) string {
	name := pane.Cell(r.Name, c.name)
	if selected {
		name = th.Accent(th.Bold(name))
	}
	line := "  "
	if selected {
		line = th.Accent(Pointer)
	}
	line += Glyph(th, r.Mark, tick) + "  " + name
	if c.kind > 0 {
		line += th.Muted(pane.Cell(r.Kind, c.kind))
	}
	if c.lang > 0 {
		line += th.Muted(pane.Cell(r.Language, c.lang))
	}
	if c.version > 0 {
		line += th.Muted(pane.Cell(r.Version, c.version))
	}
	if r.Changed {
		mark := words.Say("projectpage.changed", "dot", th.Icons.Dot)
		if c.changed == 1 {
			mark = th.Icons.Dot
		}
		line += th.Accent(mark)
	}
	line = pane.Fit(line, w)
	if selected {
		return th.Selected(line)
	}
	return line
}

// Legend is the muted line at the foot of the list: what each mark means.
//
// Parameters: th is the theme; w is the width.
// Returns: one line of w cells.
func Legend(th theme.Theme, w int) string {
	text := words.Say("projectpage.legend", "passed", th.Icons.Passed, "failed", th.Icons.Failed, "none", th.Icons.NotRun, "dot", th.Icons.Dot)
	return pane.Fit(pane.CutEllipsis(th.Muted(text), w), w)
}
