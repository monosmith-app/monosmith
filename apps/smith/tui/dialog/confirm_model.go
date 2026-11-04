package dialog

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
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
	Offer, Commit bool
	// scroll is how many lines of the head (the sentence and the details) are
	// scrolled off: a long plan never pushes the commit box and the command
	// out of the dialog.
	scroll           int
	screenW, screenH int
	// top and bottom are the rows of the screen the dialog stays clear of.
	top, bottom int
	// working is what smith is doing while the plan is read ("" once the plan
	// is shown); tick and elapsed move its spinner line (see Pulse).
	working string
	// reading is true when the command only reads (it opens a screen, it
	// changes nothing): the dialog then says the command is running.
	reading bool
	tick    int
	elapsed time.Duration
}

// NewConfirm builds the dialog for a command line (smith's arguments).
func NewConfirm(title, why string, argv []string) Confirm {
	return Confirm{Title: title, Why: why, Command: commands.Line(argv), Argv: argv, hover: -1, Offer: Committable(argv)}
}

// NewWaiting builds the dialog of a command whose plan is still being read. It
// opens at once, so the key press shows an answer: a sentence, a spinner line
// and the command that will run. Run cannot be used until the plan replaces it;
// Esc cancels the wait.
//
// Parameters: title and why are the words of the finished dialog; argv is the
// command line that will run; doing says what smith is doing now, without the
// closing dots ("Working out what will change").
// Returns: the dialog in its working state.
func NewWaiting(title, why string, argv []string, doing string) Confirm {
	c := NewConfirm(title, why, argv)
	c.Offer, c.working, c.focus = false, doing, 1
	c.Details = []string{doing + "…"}
	return c
}

// Reading marks the dialog of a command that only reads, so it says "is
// running" and not "will run".
//
// Returns: the dialog, marked.
func (c Confirm) Reading() Confirm {
	c.reading = true
	return c
}

// Working is true while the plan is still being read: Run is not available.
func (c Confirm) Working() bool { return c.working != "" }

// Pulse moves the spinner line of a dialog that waits for its plan.
//
// Parameters: tick is the animation frame; elapsed is how long the wait lasts.
// Returns: the dialog with the moving line, or c as it is when it is not waiting.
func (c Confirm) Pulse(tick int, elapsed time.Duration) Confirm {
	c.tick, c.elapsed = tick, elapsed
	return c
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
	// detailWide is the least width of a dialog that lists what a command touches.
	detailWide = 84
	buttonRun  = "[ Run  Enter ]"
	buttonStop = "[ Cancel  Esc ]"
	buttonOK   = "[ OK  Enter ]"
)

// WithDetails adds the lines that say what the command will touch.
func (c Confirm) WithDetails(lines []string) Confirm {
	c.Details = lines
	return c
}

// Within keeps the dialog off the top rows (the header and the tab bar) and the
// bottom rows (the status bar) of the screen: Place centres it in what is left.
func (c Confirm) Within(top, bottom int) Confirm {
	c.top, c.bottom = max(top, 0), max(bottom, 0)
	return c
}

// Place centres the dialog on a screen of the given size, between the margins
// Within set.
func (c Confirm) Place(sw, sh int) Confirm {
	c.screenW, c.screenH = sw, sh
	avail := max(sh-c.top-c.bottom, 3)
	c.Command = commands.Line(c.Final())
	w := min(max(ansi.StringWidth(c.Command)+8, minWidth), max(sw-2, 10))
	if c.Offer {
		// Wide enough for the box line, which is not cut.
		w = min(max(w, ansi.StringWidth(CommitLabel)+14), max(sw-2, 10))
	}
	if len(c.Details) > 0 {
		// Lines of paths need room: a short command must not make a narrow dialog
		// that breaks a path in the middle of a word.
		w = max(w, min(detailWide, max(sw-2, 10)))
	}
	if c.Info {
		w = min(noticeWide, max(sw-2, 10))
	}
	inner := w - 4
	c.cmdLines = nil
	if !c.Info {
		c.cmdLines = wrap(c.Command, inner-4) // room for the "  $ " in front
	}
	body := c.bodyLines(inner)
	h := len(body) + 4 // borders, blank line, buttons
	h = min(h, avail)
	c.rect = Rect{X: max((sw-w)/2, 0), Y: c.top + max((avail-h)/2, 0), W: w, H: h}
	return c
}

// Rect is where the dialog sits after Place.
func (c Confirm) Rect() Rect { return c.rect }

// headLines are the sentence and the details: the part that scrolls.
func (c Confirm) headLines(inner int) []string {
	body := append([]string(nil), wrap(c.Why, inner)...)
	if len(c.Details) > 0 {
		body = append(body, "")
		for _, d := range c.Details {
			body = append(body, wrap(d, inner)...)
		}
	}
	return body
}

// tailLines are the commit box and the command that will run: they always
// stay on screen. A notice has none.
func (c Confirm) tailLines() []string {
	if c.Info {
		return nil
	}
	var tail []string
	if c.Offer {
		box := "[ ] "
		if c.Commit {
			box = "[x] "
		}
		tail = append(tail, "", box+CommitLabel+"  (C)")
	}
	caption := "The command that will run:"
	if c.reading {
		caption = "The command that is running:"
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

// shownHead is the part of the head on screen, and how many lines are hidden
// below it: when the head does not fit, the last row says so.
func (c Confirm) shownHead(inner int) (shown []string, hidden int) {
	head, room := c.headLines(inner), c.headRoom()
	if len(head) <= room {
		return head, 0
	}
	if room < 2 {
		return nil, len(head)
	}
	from := min(max(c.scroll, 0), len(head)-(room-1))
	end := from + room - 1
	return head[from:end], len(head) - end
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
	rows := make([]string, 0, r.H-2)
	row := func(content string) string { return " " + pane.Fit(content, r.W-4) }
	shown, hidden := c.shownHead(inner)
	if c.Working() && hidden == 0 && len(shown) > 0 {
		shown = append(append([]string(nil), shown[:len(shown)-1]...), th.Accent(th.Spin(c.tick))+" "+progress.Waiting(c.working+"…", c.elapsed))
	}
	for _, l := range shown {
		rows = append(rows, row(l))
	}
	if hidden > 0 && c.headRoom() >= 2 {
		rows = append(rows, row(th.Muted(fmt.Sprintf("… %d more lines (Up/Down/PgDn show them)", hidden))))
	}
	for len(rows) < c.headRoom() {
		rows = append(rows, row(""))
	}
	tail := c.tailLines()
	for i, l := range tail {
		if i >= len(tail)-len(c.cmdLines) {
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
		switch {
		case c.Working() && i == 0:
			// Run waits for the plan: it is drawn faint and does nothing.
			text = th.Faint(label)
		case i == c.focus:
			text = th.Selected(label)
		case i == c.hover:
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
	return c.headRoom() + 1
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
		if c.Working() {
			return c, Outcome{}
		}
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
	if c.Working() {
		// The plan is not ready: only the way out works, so a hurried Enter
		// cannot run what the person has not seen.
		switch name {
		case "n", "N", "esc", "ctrl+c", "q":
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
	case "up", "k":
		c.scroll = max(c.scroll-1, 0)
	case "down", "j":
		c.scroll++
	case "pgup":
		c.scroll = max(c.scroll-max(c.headRoom()-1, 1), 0)
	case "pgdown":
		c.scroll += max(c.headRoom()-1, 1)
	}
	if head := len(c.headLines(max(c.rect.W-4, 1))); c.scroll > head {
		c.scroll = head
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
