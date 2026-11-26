package dashboard

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// buttonGap is the cells between two buttons of a row.
const buttonGap = 2

// WideFrom is the page width from which Home has its wide look: the four
// boxes in one row, each item with its tag and its buttons, and the row of
// Start something. A narrower page has the boxes in a 2x2 grid and shows only
// the key of each item.
const WideFrom = 88

// canvas collects the lines of the page and the click targets on them.
type canvas struct {
	th      theme.Theme
	w       int
	wide    bool
	ptr     Pointer
	ch      Choice
	tagW    int
	lines   []string
	targets []Target
	items   []Rows
}

// Page draws Home in w cells with the keyboard on the first item of the list.
//
// Parameters: th styles the page; f is what is known; w is the width; ptr is
// the cell under the mouse, so the button under it shows its hover state.
// Returns: the lines, the targets of the clicks and the lines of each item.
func Page(th theme.Theme, f Facts, w int, ptr Pointer) Rendered {
	return Render(th, f, w, ptr, Choice{})
}

// PageAt is Page with the keyboard on item sel of the list and its first
// button; a negative sel is no item, and then no button carries Enter.
//
// Parameters: as Page, and sel is the item of the list the keyboard is on.
// Returns: the lines, the targets of the clicks and the lines of each item.
func PageAt(th theme.Theme, f Facts, w int, ptr Pointer, sel int) Rendered {
	return Render(th, f, w, ptr, Choice{Item: sel})
}

// Render draws Home in w cells: the workspace and its summary on the top line,
// the four summary boxes, the list "Needs you now" with the item ch.Item
// selected, and on a wide page the buttons of Start something. A tip for a
// new person ends the page.
//
// Parameters: th styles the page; f is what is known; w is the width; ptr is
// the cell under the mouse; ch is where the keyboard is on the list. An item
// or a button past the end is taken as the last one.
// Returns: the lines, the targets of the clicks and the lines of each item.
func Render(th theme.Theme, f Facts, w int, ptr Pointer, ch Choice) Rendered {
	q := Queue(f)
	if ch.Item >= len(q) {
		ch.Item = len(q) - 1
	}
	if ch.Item >= 0 {
		ch.Button = min(max(ch.Button, 0), max(len(q[ch.Item].Buttons)-1, 0))
	}
	c := &canvas{th: th, w: w, wide: w >= WideFrom, ptr: ptr, ch: ch}
	c.add("")
	c.summary(f)
	c.add("")
	c.boxes(f)
	c.add("")
	c.list(q)
	if c.wide && q[0].Kind != KindEmpty {
		c.add("", " "+th.Border(strings.Repeat("─", max(w-2, 0)), false), "")
		c.add(" " + th.Bold(words.Say("dashboard.start.title")))
		c.add("")
		c.buttons(startButtons(), 1)
	}
	if f.SettingsTip {
		tip := words.Say("dashboard.settings.tip", "new", th.Accent(words.Say("dashboard.settings.new")), "shortcut", commands.ShortcutOf("view.settings"))
		c.add("")
		c.add(pane.Wrap(" "+th.Muted(tip), w, 1)...)
	}
	c.add("")
	for i := range c.lines {
		c.lines[i] = pane.Fit(c.lines[i], w)
	}
	return Rendered{Lines: c.lines, Targets: c.targets, Items: c.items}
}

func (c *canvas) add(lines ...string) { c.lines = append(c.lines, lines...) }

// summary is the top line: the workspace in bold, then how many projects it
// holds, how many changed since the base and when the last run ended. A part
// without facts is left out, and the languages go when the line is too long.
func (c *canvas) summary(f Facts) {
	th := c.th
	parts := func(projects string) string {
		var out []string
		if f.ProjectsLoaded {
			out = append(out, projects)
			if f.Projects > 0 {
				out = append(out, changedText(f))
			}
		}
		if len(f.Recent) > 0 && f.Recent[0].Ago != "" {
			out = append(out, words.Say("dashboard.top.last_run", "ago", f.Recent[0].Ago))
		}
		return strings.Join(out, " · ")
	}
	lead := " "
	if f.Workspace != "" {
		lead += th.Bold(f.Workspace) + "  "
	}
	rest := parts(projectsText(f))
	if pane.Width(lead)+pane.Width(rest) > c.w-1 {
		rest = parts(words.Count("dashboard.project", f.Projects))
	}
	c.add(lead + th.Muted(pane.CutEllipsis(rest, max(c.w-1-pane.Width(lead), 1))))
}

// changedText says how many projects changed, and since when when the base is known.
func changedText(f Facts) string {
	if f.Base == "" {
		return words.Say("dashboard.top.changed", "count", f.Affected)
	}
	return words.Say("dashboard.top.changed_since", "count", f.Affected, "base", f.Base)
}

// startButtons are the common actions Home offers under the list.
func startButtons() []Button {
	return []Button{
		queueButton("dashboard.good.build", "run.build-all"),
		queueButton("dashboard.start.lint", "run.lint-all"),
		// The palette has three keys; the button names the one the title bar shows.
		{Label: words.Say("dashboard.start.task"), CommandID: "file.palette", Key: "ctrl+k"},
		queueButton("dashboard.add.label", "file.add"),
		queueButton("dashboard.compose.label", "file.compose"),
		queueButton("dashboard.start.settings", "view.settings"),
	}
}

// icon is the status mark of a severity, in its colour.
func (c *canvas) icon(s Severity) string {
	th := c.th
	switch s {
	case SeverityFailure:
		return th.Failure(th.Icons.Failed)
	case SeverityWarning:
		return th.Warning(th.Icons.Warning)
	case SeveritySuccess:
		return th.Success(th.Icons.Passed)
	case SeverityInfo:
		return th.Accent(th.Icons.Dot)
	}
	return th.Muted(th.Icons.Queued)
}

// paint draws s in the colour of a severity.
func (c *canvas) paint(s Severity, text string) string {
	th := c.th
	switch s {
	case SeverityFailure:
		return th.Failure(text)
	case SeverityWarning:
		return th.Warning(text)
	case SeveritySuccess:
		return th.Success(text)
	case SeverityInfo:
		return th.Accent(text)
	}
	return th.Muted(text)
}

// buttons lays buttons out from column x on a new line, and on the next line
// when one does not fit. They belong to no item of the list.
func (c *canvas) buttons(bs []Button, x0 int) {
	row := len(c.lines)
	line, x := strings.Repeat(" ", x0), x0
	for i, b := range bs {
		key := c.keyOf(b, false)
		bw := button.Width(c.th, b.Label, key)
		if i > 0 && x+buttonGap+bw > c.w {
			c.add(line)
			row++
			line, x = strings.Repeat(" ", x0), x0
		} else if i > 0 {
			line += strings.Repeat(" ", buttonGap)
			x += buttonGap
		}
		line += c.draw(b, key, row, x, false, -1, i)
		x += bw
	}
	c.add(line)
}

// draw is a button at column x of line row: primary when active, in its hover
// state under the pointer, else secondary. It records where a click lands.
func (c *canvas) draw(b Button, key string, row, x int, active bool, item, place int) string {
	bw := button.Width(c.th, b.Label, key)
	state := button.Rest
	switch {
	case active:
		state = button.Active
	case c.ptr.Row == row && c.ptr.Col >= x && c.ptr.Col < x+bw:
		state = button.Hover
	}
	c.targets = append(c.targets, Target{Row: row, From: x, To: x + bw, Command: b.CommandID, Kind: TargetButton, Item: item, Button: place})
	return button.Render(c.th, b.Label, key, state)
}

// keyOf is the key a button shows: Enter while the keyboard is on it, else
// its own key. The first button of the queue carries Enter in its Key, and
// gets the shortcut of its command back.
func (c *canvas) keyOf(b Button, chosen bool) string {
	enter := commands.ShortcutOf("home.next")
	switch {
	case chosen:
		return enter
	case b.Key == enter:
		return commands.ShortcutOf(b.CommandID)
	}
	return b.Key
}
