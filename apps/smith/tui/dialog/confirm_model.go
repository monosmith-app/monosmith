package dialog

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Confirm asks before a command that changes the workspace or publishes. It
// shows the exact command line that will run, so what is confirmed is what
// runs. Keys: Enter or Y runs, Esc or N cancels, Tab and the arrows move
// between the buttons.
type Confirm struct {
	Title   string
	Why     string
	Command string
	// Details are lines between the sentence and the command: what the
	// command will touch (the files, the packages and versions).
	Details []string
	// Argv is the command line without the program name; the shell runs it on confirmation.
	Argv []string
	// Info makes the dialog a notice: it only tells (Why and Details), has an
	// OK button and runs nothing.
	Info bool
	// focus is 0 for Run, 1 for Cancel; hover is the button under the pointer or -1.
	focus, hover int
	rect         Rect
	cmdLines     []string
	// Offer shows the box "Commit this change so it can be undone"; Commit is
	// true once it is ticked. It starts unticked: smith commits only when asked.
	Offer, Commit    bool
	screenW, screenH int
}

// NewConfirm builds the dialog for a command line (smith's arguments).
func NewConfirm(title, why string, argv []string) Confirm {
	return Confirm{Title: title, Why: why, Command: commands.Line(argv), Argv: argv, hover: -1, Offer: Committable(argv)}
}

// Final is the command line that runs on confirmation: Argv, with --commit
// when the box is ticked.
func (c Confirm) Final() []string { return WithCommit(c.Argv, c.Offer && c.Commit) }

// toggleCommit ticks or unticks the box and redraws the command line.
func (c Confirm) toggleCommit() Confirm {
	c.Commit = !c.Commit
	return c.Place(c.screenW, c.screenH)
}

// NewNotice builds a dialog that only tells: a title, a sentence (Why) and
// the lines of Details, closed with OK. It is how the shell shows the error of
// a command, with the fix the command named.
func NewNotice(title, why string, details []string) Confirm {
	return Confirm{Title: title, Why: why, Details: details, Info: true, hover: -1}
}

const (
	minWidth   = 44
	noticeWide = 60
	buttonRun  = "[ Run  Enter ]"
	buttonStop = "[ Cancel  Esc ]"
	buttonOK   = "[ OK  Enter ]"
)

// WithDetails adds the lines that say what the command will touch.
func (c Confirm) WithDetails(lines []string) Confirm {
	c.Details = lines
	return c
}

// Place centres the dialog on a screen of the given size.
func (c Confirm) Place(sw, sh int) Confirm {
	c.screenW, c.screenH = sw, sh
	c.Command = commands.Line(c.Final())
	w := min(max(ansi.StringWidth(c.Command)+8, minWidth), max(sw-2, 10))
	if c.Offer {
		// Wide enough for the box line, which is not cut.
		w = min(max(w, ansi.StringWidth(CommitLabel)+14), max(sw-2, 10))
	}
	if c.Info {
		w = min(noticeWide, max(sw-2, 10))
	}
	inner := w - 4
	c.cmdLines = nil
	if !c.Info {
		c.cmdLines = wrap(c.Command, inner-2)
	}
	body := c.bodyLines(inner)
	h := len(body) + 4 // borders, blank line, buttons
	h = min(h, max(sh, 3))
	c.rect = Rect{X: max((sw-w)/2, 0), Y: max((sh-h)/2, 0), W: w, H: h}
	return c
}

// Rect is where the dialog sits after Place.
func (c Confirm) Rect() Rect { return c.rect }

func (c Confirm) bodyLines(inner int) []string {
	body := append([]string(nil), wrap(c.Why, inner)...)
	if len(c.Details) > 0 {
		body = append(body, "")
		for _, d := range c.Details {
			body = append(body, wrap(d, inner)...)
		}
	}
	if c.Info {
		return body
	}
	if c.Offer {
		box := "[ ] "
		if c.Commit {
			box = "[x] "
		}
		body = append(body, "", box+CommitLabel+"  (C)")
	}
	body = append(body, "", "The command that will run:")
	for _, l := range c.cmdLines {
		body = append(body, "  $ "+l)
	}
	return body
}

// buttonsRow returns the row of buttons for the inner width and the columns
// (relative to the dialog's left edge) of each.
func (c Confirm) buttonsRow(inner int) (spans [2][2]int) {
	run, stop := ansi.StringWidth(buttonRun), ansi.StringWidth(buttonStop)
	if c.Info {
		run, stop = ansi.StringWidth(buttonOK), 0
	}
	total := run + 2 + stop
	start := 2 + max((inner-total)/2, 0)
	spans[0] = [2]int{start, start + run}
	spans[1] = [2]int{start + run + 2, start + run + 2 + stop}
	return spans
}

// Lines draws the dialog: Rect().H lines of Rect().W cells, in the same rounded border as a pane, in the accent colour.
func (c Confirm) Lines(th theme.Theme) []string {
	r := c.rect
	if r.W < 6 || r.H < 3 {
		return nil
	}
	inner := r.W - 4
	body := c.bodyLines(inner)
	rows := make([]string, 0, r.H-2)
	row := func(content string) string { return " " + pane.Fit(content, r.W-4) }
	for i, l := range body {
		if len(rows) >= r.H-4 {
			break
		}
		if i >= len(body)-len(c.cmdLines) {
			l = th.Bold(l)
		}
		rows = append(rows, row(l))
	}
	rows = append(rows, row(""))
	spans := c.buttonsRow(inner)
	line := strings.Repeat(" ", spans[0][0]-2)
	labels := []string{buttonRun, buttonStop}
	if c.Info {
		labels = []string{buttonOK}
	}
	for i, label := range labels {
		text := th.Muted(label)
		switch i {
		case c.focus:
			text = th.Selected(label)
		case c.hover:
			text = th.Hover(label)
		}
		line += text
		if i == 0 && len(labels) > 1 {
			line += "  "
		}
	}
	rows = append(rows, row(line))
	return pane.Frame(th, c.Title, rows, r.W, r.H, true)
}

func (c Confirm) buttonAt(x, y int) int {
	r := c.rect
	if y != r.Y+r.H-2 {
		return -1
	}
	spans := c.buttonsRow(r.W - 4)
	for i, s := range spans {
		if x-r.X >= s[0] && x-r.X < s[1] {
			return i
		}
	}
	return -1
}

// commitRow is the index, in the body, of the line with the box.
func (c Confirm) commitRow() int {
	inner := max(c.rect.W-4, 1)
	return len(c.bodyLines(inner)) - len(c.cmdLines) - 3
}

// Hover highlights the button under the pointer.
func (c Confirm) Hover(x, y int) Confirm { c.hover = c.buttonAt(x, y); return c }

// Click handles a left press: a button runs or cancels; a press outside the
// dialog cancels; a press inside elsewhere does nothing.
func (c Confirm) Click(x, y int) (Confirm, Outcome) {
	if c.Offer && c.rect.Contains(x, y) && y == c.rect.Y+1+c.commitRow() {
		return c.toggleCommit(), Outcome{}
	}
	switch c.buttonAt(x, y) {
	case 0:
		return c, Outcome{Done: true, Confirmed: !c.Info}
	case 1:
		return c, Outcome{Done: true}
	}
	if !c.rect.Contains(x, y) {
		return c, Outcome{Done: true}
	}
	return c, Outcome{}
}

// Key handles a key name.
func (c Confirm) Key(name string) (Confirm, Outcome) {
	if c.Info {
		switch name {
		case "enter", " ", "esc", "q", "y", "n", "ctrl+c":
			return c, Outcome{Done: true}
		}
		return c, Outcome{}
	}
	switch name {
	case "c", "C":
		if c.Offer {
			return c.toggleCommit(), Outcome{}
		}
	case "y", "Y":
		return c, Outcome{Done: true, Confirmed: true}
	case "n", "N", "esc", "ctrl+c", "q":
		return c, Outcome{Done: true}
	case "enter", " ":
		return c, Outcome{Done: true, Confirmed: c.focus == 0}
	case "tab", "right", "left", "shift+tab":
		c.focus = 1 - c.focus
	}
	return c, Outcome{}
}

// wrap breaks s into lines of at most w cells at spaces, hard-cutting words
// that are longer.
func wrap(s string, w int) []string {
	if w < 1 {
		return []string{s}
	}
	var out []string
	for _, para := range strings.Split(s, "\n") {
		line := ""
		for _, word := range strings.Fields(para) {
			for ansi.StringWidth(word) > w {
				if line != "" {
					out, line = append(out, line), ""
				}
				out = append(out, ansi.Truncate(word, w, ""))
				word = ansi.TruncateLeft(word, w, "")
			}
			switch {
			case line == "":
				line = word
			case ansi.StringWidth(line)+1+ansi.StringWidth(word) <= w:
				line += " " + word
			default:
				out, line = append(out, line), word
			}
		}
		out = append(out, line)
	}
	return out
}
