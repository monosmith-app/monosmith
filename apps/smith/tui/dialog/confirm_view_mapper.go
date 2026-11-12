package dialog

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// headLines are the sentence and the details: the part that scrolls.
func (c Confirm) headLines(inner int) []string {
	body := append([]string(nil), pane.WrapWords(c.Why, inner)...)
	if len(c.Details) > 0 {
		body = append(body, "")
		for _, d := range c.Details {
			body = append(body, pane.WrapWords(d, inner)...)
		}
	}
	return body
}

// tailLines are the commit box and the command that will run: they always
// stay on screen. A notice has none.
func (c Confirm) tailLines() []string {
	return c.tailLinesIn(theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII}))
}

// tailLinesIn is tailLines drawn in a theme: the commit box shows the row
// hover look in it while the pointer is over the box.
func (c Confirm) tailLinesIn(th theme.Theme) []string {
	if c.Info {
		return nil
	}
	var tail []string
	if c.Offer {
		tail = append(tail, "", choicerow.Toggle(th, CommitLabel+words.Say("dialog.commit.key"), c.Commit, c.overCommit))
	}
	caption := words.Say("dialog.command.will")
	if c.reading {
		caption = words.Say("dialog.command.running")
	}
	tail = append(tail, "", caption)
	for i, l := range c.cmdLines {
		// Only the first line has the prompt: a "$" on each would read as many commands.
		prefix := "  $ "
		if i > 0 {
			prefix = "    "
		}
		tail = append(tail, prefix+l)
	}
	return tail
}

func (c Confirm) bodyLines(inner int) []string {
	return append(c.headLines(inner), c.tailLines()...)
}

// headRoom is how many rows the head has on screen.
func (c Confirm) headRoom() int { return max(c.rect.H-4-len(c.tailLines()), 0) }

// Lines draws the dialog: Rect().H lines of Rect().W cells, in the same rounded border as a pane, in the accent colour.
func (c Confirm) Lines(th theme.Theme) []string {
	r := c.rect
	if r.W < 6 || r.H < 3 {
		return nil
	}
	inner := r.W - 4
	rows := make([]string, 0, r.H-2)
	row := func(content string) string { return " " + pane.Fit(content, r.W-4) }
	head := c.headLines(inner)
	if c.Working() && len(head) > 0 && len(head) <= c.headRoom() {
		head = append(append([]string(nil), head[:len(head)-1]...), th.Accent(th.Spin(c.tick))+" "+progress.Waiting(c.working+"…", c.elapsed))
	}
	// The head scrolls under the shared scrollbar and "more" rows; the commit
	// box, the command and the buttons stay under it.
	for _, l := range pane.DialogBody(th, head, nil, inner, c.headRoom(), c.scroll) {
		rows = append(rows, row(l))
	}
	tail := c.tailLinesIn(th)
	for i, l := range tail {
		if i >= len(tail)-len(c.cmdLines) {
			l = th.Bold(l)
		}
		rows = append(rows, row(l))
	}
	rows = append(rows, row(""))
	labels, keys := c.buttonLabels()
	spans := button.RowSpans(th, labels, keys, inner)
	line := strings.Repeat(" ", spans[0][0]-2)
	for i, label := range labels {
		state := button.Rest
		switch {
		case c.Working() && i == 0:
			// Run waits for the plan: it is drawn disabled and does nothing.
			state = button.Disabled
		case i == c.hover:
			state = button.Hover
		case i == c.focus:
			state = button.Active
		}
		line += button.Render(th, label, keys[i], state)
		if i == 0 && len(labels) > 1 {
			line += " "
		}
	}
	rows = append(rows, row(line))
	return pane.Frame(th, c.Title, rows, r.W, r.H, true)
}
