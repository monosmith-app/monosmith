package dialog

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
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
	// overCommit is true while the pointer is over the commit box.
	overCommit bool
	rect       layout.Rect
	cmdLines   []string
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
	// runLabel replaces the label of the run button; line replaces the command
	// line drawn (a command that is not a smith command line).
	runLabel, line string
	// cancelFirst is true when Cancel starts focused: Enter then cancels, so
	// the run button shows the key y, which confirms.
	cancelFirst bool
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

// WithRunLabel names the button that confirms, for a dialog whose action has a
// name of its own ("Release now").
//
// Parameters: label is the text of the button.
// Returns: the dialog with the new label.
func (c Confirm) WithRunLabel(label string) Confirm {
	c.runLabel = label
	return c
}

// FocusCancel puts the focus on Cancel, so a hurried Enter cancels: the dialog
// of a step that cannot be undone starts there.
//
// Returns: the dialog with Cancel focused.
func (c Confirm) FocusCancel() Confirm {
	c.focus, c.cancelFirst = 1, true
	return c
}

// WithLine sets the command line the dialog draws, for a command that is not a
// smith command line (it runs git, for example). Argv still tells the shell what
// was confirmed.
//
// Parameters: line is the text drawn after the "$".
// Returns: the dialog with the line.
func (c Confirm) WithLine(line string) Confirm {
	c.line = line
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
)

// The labels of the buttons, and the keys that do what a click does.
var (
	buttonRun  = words.Say("dialog.button.run")
	buttonStop = words.Say("dialog.button.cancel")
	buttonOK   = words.Say("dialog.button.ok")
)

const (
	keyEnter = "enter" // words:literal
	keyEsc   = "esc"   // words:literal
	keyYes   = "y"     // words:literal
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
	if c.line != "" {
		c.Command = c.line
	}
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
		c.cmdLines = pane.WrapWords(c.Command, inner-4) // room for the "  $ " in front
	}
	body := c.bodyLines(inner)
	h := len(body) + 4 // borders, blank line, buttons
	h = min(h, avail)
	c.rect = layout.Rect{X: max((sw-w)/2, 0), Y: c.top + max((avail-h)/2, 0), W: w, H: h}
	return c
}

// Rect is where the dialog sits after Place.
func (c Confirm) Rect() layout.Rect { return c.rect }

// buttonLabels are the labels and keys of the buttons: Run and Stop, or only
// OK for an info dialog.
func (c Confirm) buttonLabels() (labels, keys []string) {
	if c.Info {
		return []string{buttonOK}, []string{keyEnter}
	}
	run := buttonRun
	if c.runLabel != "" {
		run = c.runLabel
	}
	if c.cancelFirst {
		return []string{run, buttonStop}, []string{keyYes, keyEsc}
	}
	return []string{run, buttonStop}, []string{keyEnter, keyEsc}
}

// buttonsRow returns the columns (relative to the dialog's left edge) of each
// button for the inner width.
func (c Confirm) buttonsRow(inner int) [][2]int {
	labels, keys := c.buttonLabels()
	return button.RowSpans(theme.Default(), labels, keys, inner)
}

// commitRow is the index, in the body, of the line with the box.
func (c Confirm) commitRow() int {
	return c.headRoom() + 1
}

// Hover highlights the button under the pointer.
func (c Confirm) Hover(x, y int) Confirm {
	c.hover = button.At(c.buttonsRow(c.rect.W-4), c.rect, x, y)
	c.overCommit = c.Offer && c.rect.Contains(x, y) && y == c.rect.Y+1+c.commitRow()
	return c
}

// Click handles a left press: a button runs or cancels; a press outside the
// dialog cancels; a press inside elsewhere does nothing.
func (c Confirm) Click(x, y int) (Confirm, Outcome) {
	if c.Offer && c.rect.Contains(x, y) && y == c.rect.Y+1+c.commitRow() {
		return c.toggleCommit(), Outcome{}
	}
	switch button.At(c.buttonsRow(c.rect.W-4), c.rect, x, y) {
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
		case "n", "N", "esc", "ctrl+c", "q": // words:literal
			return c, Outcome{Done: true}
		}
		return c, Outcome{}
	}
	page := pane.PageStep(c.headRoom())
	switch commands.Nav(name) {
	case commands.NavUp:
		return c.Wheel(-1), Outcome{}
	case commands.NavDown:
		return c.Wheel(1), Outcome{}
	case commands.NavPageUp:
		return c.Wheel(-page), Outcome{}
	case commands.NavPageDown:
		return c.Wheel(page), Outcome{}
	default:
		switch name {
		case "ctrl+u":
			return c.Wheel(-page), Outcome{}
		case "ctrl+d":
			return c.Wheel(page), Outcome{}
		case "c", "C": // words:literal
			if c.Offer {
				return c.toggleCommit(), Outcome{}
			}
		case "y", "Y": // words:literal
			return c, Outcome{Done: true, Confirmed: true}
		case "n", "N", "esc", "ctrl+c", "q": // words:literal
			return c, Outcome{Done: true}
		case "enter", " ":
			return c, Outcome{Done: true, Confirmed: c.focus == 0}
		case "tab", "right", "left", "shift+tab":
			c.focus = 1 - c.focus
		}
	}
	return c, Outcome{}
}

// Wheel scrolls the head (the sentence and the details) by delta lines, never
// past its ends; the commit box, the command and the buttons stay.
//
// Parameters: delta is the lines to move, negative for up.
// Returns: the dialog with the head scrolled.
func (c Confirm) Wheel(delta int) Confirm {
	c.scroll = pane.ClampTop(c.scroll+delta, len(c.headLines(max(c.rect.W-4, 1))), c.headRoom())
	return c
}
