package dashboard

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// buttonGap is the cells between two buttons of a row.
const buttonGap = 2

// canvas collects the lines of the page and the click targets on them.
type canvas struct {
	th      theme.Theme
	w       int
	ptr     Pointer
	lines   []string
	targets []Target
}

// Page draws the attention queue in w cells: every item the same way (its
// headline, an optional note line and its buttons on a line of their own,
// left-aligned under the headline), then the recent runs, and the workspace
// with the buttons that add to it.
//
// Parameters: th styles the page; f is what is known; w is the width; ptr is
// the cell under the mouse, so the button under it shows its hover state.
// Returns: the lines and the targets of the clicks.
func Page(th theme.Theme, f Facts, w int, ptr Pointer) Rendered {
	c := &canvas{th: th, w: w, ptr: ptr}
	q := Queue(f)
	c.add("")
	for _, it := range q {
		c.item(it)
		c.add("")
	}
	if len(f.Recent) > 0 && (len(q) != 1 || q[0].Kind != KindChecking) {
		c.recent(f.Recent)
		c.add("")
	}
	if f.SettingsTip {
		tip := words.Say("dashboard.settings.tip", "new", th.Accent(words.Say("dashboard.settings.new")), "shortcut", commands.ShortcutOf("view.settings"))
		c.add(pane.Wrap(" "+th.Muted(tip), w, 1)...)
		c.add("")
	}
	if f.ProjectsLoaded && f.Projects > 0 {
		c.add(pane.Wrap(" "+th.Muted(workspaceText(f)), w, 1)...)
	}
	if q[0].Kind != KindEmpty {
		c.buttons([]Button{queueButton("dashboard.add.label", "file.add"), queueButton("dashboard.compose.label", "file.compose")}, 1)
	}
	for i := range c.lines {
		c.lines[i] = pane.Fit(c.lines[i], w)
	}
	return Rendered{Lines: c.lines, Targets: c.targets}
}

func (c *canvas) add(lines ...string) { c.lines = append(c.lines, lines...) }

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

// item draws an item: the headline in bold, the note under it and the buttons
// under that, all starting at the same column.
func (c *canvas) item(it Item) {
	lead := " " + c.icon(it.Severity) + " "
	indent := pane.Width(lead)
	c.add(pane.Wrap(lead+c.th.Bold(it.Title), c.w, indent)...)
	if it.Detail != "" {
		c.add(pane.Wrap(strings.Repeat(" ", indent)+c.th.Muted(it.Detail), c.w, indent)...)
	}
	if len(it.Buttons) > 0 {
		c.buttons(it.Buttons, indent)
	}
}

// buttons lays buttons out from column x on a new line, and on the next line
// when one does not fit.
func (c *canvas) buttons(bs []Button, x0 int) {
	row := len(c.lines)
	line, x := strings.Repeat(" ", x0), x0
	for i, b := range bs {
		bw := button.Width(c.th, b.Label, b.Key)
		if i > 0 && x+buttonGap+bw > c.w {
			c.add(line)
			row++
			line, x = strings.Repeat(" ", x0), x0
		} else if i > 0 {
			line += strings.Repeat(" ", buttonGap)
			x += buttonGap
		}
		line += c.draw(b, row, x)
		x += bw
	}
	c.add(line)
}

// draw is a button at column x of line row, in the state the pointer gives it,
// and records where a click on it lands.
func (c *canvas) draw(b Button, row, x int) string {
	bw := button.Width(c.th, b.Label, b.Key)
	c.targets = append(c.targets, Target{Row: row, From: x, To: x + bw, Command: b.CommandID})
	state := button.Rest
	if c.ptr.Row == row && c.ptr.Col >= x && c.ptr.Col < x+bw {
		state = button.Hover
	}
	return button.Render(c.th, b.Label, b.Key, state)
}
