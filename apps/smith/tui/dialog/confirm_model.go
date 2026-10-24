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
	// focus is 0 for Run, 1 for Cancel; hover is the button under the pointer or -1.
	focus, hover int
	rect         Rect
	cmdLines     []string
}

// NewConfirm builds the dialog for a command line (smith's arguments).
func NewConfirm(title, why string, argv []string) Confirm {
	return Confirm{Title: title, Why: why, Command: commands.Line(argv), Argv: argv, hover: -1}
}

const (
	minWidth   = 44
	buttonRun  = "[ Run  Enter ]"
	buttonStop = "[ Cancel  Esc ]"
)

// WithDetails adds the lines that say what the command will touch.
func (c Confirm) WithDetails(lines []string) Confirm {
	c.Details = lines
	return c
}

// Place centres the dialog on a screen of the given size.
func (c Confirm) Place(sw, sh int) Confirm {
	w := min(max(ansi.StringWidth(c.Command)+8, minWidth), max(sw-2, 10))
	inner := w - 4
	c.cmdLines = wrap(c.Command, inner-2)
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
	total := run + 2 + stop
	start := 2 + max((inner-total)/2, 0)
	spans[0] = [2]int{start, start + run}
	spans[1] = [2]int{start + run + 2, start + run + 2 + stop}
	return spans
}

// Lines draws the dialog: Rect().H lines of Rect().W cells, in a double border.
func (c Confirm) Lines(th theme.Theme) []string {
	r := c.rect
	if r.W < 6 || r.H < 3 {
		return nil
	}
	inner := r.W - 4
	body := c.bodyLines(inner)
	edge := func(s string) string { return th.Accent(s) }
	title := " " + c.Title + " "
	top := edge("╔═") + th.Title(title) + edge(strings.Repeat("═", max(r.W-3-ansi.StringWidth(title), 0))+"╗")
	rows := []string{pane.Fit(top, r.W)}
	row := func(content string) string {
		return edge("║") + " " + pane.Fit(content, r.W-4) + " " + edge("║")
	}
	for i, l := range body {
		if len(rows) >= r.H-3 {
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
	for i, label := range []string{buttonRun, buttonStop} {
		text := th.Muted(label)
		switch i {
		case c.focus:
			text = th.Selected(label)
		case c.hover:
			text = th.Hover(label)
		}
		line += text
		if i == 0 {
			line += "  "
		}
	}
	rows = append(rows, row(line))
	rows = append(rows, edge("╚"+strings.Repeat("═", r.W-2)+"╝"))
	return pane.Block(rows, r.W, r.H)
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

// Hover highlights the button under the pointer.
func (c Confirm) Hover(x, y int) Confirm { c.hover = c.buttonAt(x, y); return c }

// Click handles a left press: a button runs or cancels; a press outside the
// dialog cancels; a press inside elsewhere does nothing.
func (c Confirm) Click(x, y int) (Confirm, Outcome) {
	switch c.buttonAt(x, y) {
	case 0:
		return c, Outcome{Done: true, Confirmed: true}
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
	switch name {
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
