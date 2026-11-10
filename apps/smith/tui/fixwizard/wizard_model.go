package fixwizard

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Show writes a fix as the review shows it: its description, the diff of each
// file, the commands (doctorfix.ShowFix against the workspace root).
type Show func(doctorfix.Planned) (string, error)

// Wizard walks one planned fix. Keys: Up and Down (or j and k) move between
// the options, a digit picks one, Enter chooses; in the review Enter applies
// and the arrows and Page keys scroll; Esc cancels (while a step runs, it
// stops the step, as Ctrl+C does); b or Left goes back one question.
type Wizard struct {
	// Fix is the planned fix being walked, before any answer.
	Fix   doctorfix.Planned
	show  Show
	style Style
	// busy is the headline of a step that was started with Start.
	busy string

	phase    Phase
	answers  []int
	resolved protocol.Fix
	question *protocol.FixChoice
	// cursor is the option under the cursor while asking.
	cursor int
	// text is the editable field of Style.Input; editing says it has the keys.
	text    string
	editing bool
	review  []string
	// summary is review with each diff cut to one line; full says the diffs show.
	summary []string
	hasDiff bool
	full    bool
	// top is the first visible line of the body.
	top int

	step, total int
	// started is when the applying phase began, for the elapsed time.
	started  time.Time
	stepName string
	headline string
	output   []string
	// note is the last lines a step printed (at most NoteRows), oldest first.
	note []string
	// log is everything the step printed so far; logOpen says the full log
	// view shows it, and tailing that the view stays on the newest line.
	log     []string
	logOpen bool
	tailing bool

	sw, sh int
	hover  int
}

// New starts the walk of a planned fix on its first question, or on the
// review when it asks nothing. show draws the review.
func New(p doctorfix.Planned, show Show) Wizard {
	w := Wizard{Fix: p, show: show, hover: -1}
	return w.resolve()
}

// Phase is where the walk is.
func (w Wizard) Phase() Phase { return w.phase }

// Styled words the dialog for another walk than a doctor fix.
func (w Wizard) Styled(s Style) Wizard {
	w.style = s
	return w
}

// Answers are the options picked so far, counted from 1.
func (w Wizard) Answers() []int { return append([]int(nil), w.answers...) }

// Start shows a step that is running, not asked for by the review: headline
// says what it does and total is how many steps will report (0 when they do
// not). Progress and Fail work as after Outcome.Apply.
func (w Wizard) Start(headline string, total int) Wizard {
	w = w.begin()
	w.busy, w.total = headline, total
	return w
}

// Resolved is the fix with the answers' files and steps added: what Apply runs.
func (w Wizard) Resolved() protocol.Fix { return w.resolved }

// Place tells the wizard the screen size; the dialog is centred on it.
func (w Wizard) Place(sw, sh int) Wizard {
	w.sw, w.sh = sw, sh
	return w
}

// resolve answers the questions with the answers so far and moves to the
// next question, or to the review once none is left.
func (w Wizard) resolve() Wizard {
	fix, q, err := doctorfix.Resolve(w.Fix.Fix, w.answers)
	for err != nil && len(w.answers) > 0 {
		w.answers = w.answers[:len(w.answers)-1]
		fix, q, err = doctorfix.Resolve(w.Fix.Fix, w.answers)
	}
	w.resolved, w.question = fix, q
	w.cursor, w.top = 0, 0
	if q != nil {
		w.phase = Asking
		return w
	}
	w.phase = Reviewing
	planned := w.Fix
	planned.Fix = fix
	w.review, w.summary, w.hasDiff, w.full = nil, nil, false, false
	if w.show != nil {
		text, err := w.show(planned)
		if err != nil {
			text = words.Say("fixwizard.review.failed", "description", fix.Description, "error", err.Error())
		}
		w.review = strings.Split(strings.TrimRight(text, "\n"), "\n")
		w.summary, w.hasDiff = summarizeDiffs(w.review)
	}
	return w
}

// canGoBack is true once a question was answered.
func (w Wizard) canGoBack() bool {
	return len(w.answers) > 0 && (w.phase == Asking || w.phase == Reviewing)
}

// back undoes the last answer.
func (w Wizard) back() Wizard {
	last := w.answers[len(w.answers)-1]
	w.answers = w.answers[:len(w.answers)-1]
	w = w.resolve()
	w.cursor = last - 1
	return w.follow()
}

// Progress records the step the fix is on: n of total, and what it does.
func (w Wizard) Progress(n, total int, name string) Wizard {
	if w.phase == Applying {
		w.step, w.total, w.stepName = n, total, name
	}
	return w
}

// NoteRows is how many of the last lines a running step printed the dialog
// shows under the spinner. The rows are always there, so the dialog keeps its
// height.
const NoteRows = 6

// Noted sets the last lines of detail of the running step, shown under the
// spinner: a file being written, or what a command printed. Only the last
// NoteRows lines are kept.
func (w Wizard) Noted(lines []string) Wizard {
	if len(lines) > NoteRows {
		lines = lines[len(lines)-NoteRows:]
	}
	w.note = lines
	return w
}

// Logged sets the whole log of the running step. The shell gives it before a
// key, a click or a frame while LogOpen is true. A view that follows the
// output stays on the newest line.
func (w Wizard) Logged(lines []string) Wizard {
	w.log = lines
	if w.tailing {
		w.top = w.bottom()
	}
	return w
}

// LogOpen is true while the full log of the running step shows.
func (w Wizard) LogOpen() bool { return w.phase == Applying && w.logOpen }

// bottom is the first line of the log view that shows the newest line.
func (w Wizard) bottom() int { return pane.ClampTop(1<<30, len(w.log), w.visible()) }

// Begin switches to the applying phase, as Outcome.Apply asks the shell to run the fix.
func (w Wizard) begin() Wizard {
	w.phase, w.step, w.total, w.stepName, w.top = Applying, 0, len(w.resolved.Steps), "", 0
	w.headline, w.output, w.note = "", nil, nil
	w.log, w.logOpen, w.tailing = nil, false, false
	w.started = time.Now()
	return w
}

// Fail shows how the fix stopped and offers Retry and Close.
func (w Wizard) Fail(err error) Wizard {
	w.phase, w.top = Failed, 0
	w.headline, w.output = failureOf(err, w.style.Name)
	return w
}

// FailVerify shows a fix whose steps all passed but whose check still fails:
// the finding doctor reads after the fix, and what the steps printed, so the
// person sees why (a package manager that exited 0 but did not install).
//
// Parameters: finding is the finding that survived the fix, in words; output
// is the lines the steps printed, oldest first.
func (w Wizard) FailVerify(finding string, output []string) Wizard {
	w.phase, w.top = Failed, 0
	w.headline, w.output = words.Say("fixwizard.failed.verify", "finding", finding), output
	return w
}

// FailureText is the failure as plain text, for the clipboard.
func (w Wizard) FailureText() string {
	return strings.TrimRight(w.headline+"\n"+strings.Join(w.output, "\n"), "\n")
}

// Key handles a key name.
func (w Wizard) Key(name string) (Wizard, Outcome) {
	switch w.phase {
	case Asking:
		return w.askingKey(name)
	case Reviewing:
		return w.reviewKey(name)
	case Applying:
		return w.applyingKey(name)
	case Failed:
		return w.failedKey(name)
	}
	return w, Outcome{}
}

// applyingKey handles a key while a step runs. Ctrl+C stops it. Esc closes the
// full log when it is open, and stops the step when it is not. The key l opens
// and closes the log; the arrows and Page keys scroll it. The step keeps running.
func (w Wizard) applyingKey(name string) (Wizard, Outcome) {
	switch {
	case name == "ctrl+c":
		return w, Outcome{Stop: true}
	case !w.logOpen && name == "esc":
		return w, Outcome{Stop: true}
	case !w.logOpen && name == "l":
		w.logOpen, w.tailing = true, true
		w.top = w.bottom()
	case w.logOpen && (name == "esc" || name == "l" || name == "q"):
		w.logOpen, w.top = false, 0
	case w.logOpen:
		w = w.scrollKey(name)
	}
	return w, Outcome{}
}

// maxText is the longest text the editable field takes.
const maxText = 60

// Text is the text of the editable field (Style.Input).
func (w Wizard) Text() string { return w.text }

// WithText sets the text of the editable field.
func (w Wizard) WithText(text string) Wizard {
	w.text = text
	return w
}

// Editing is true while the editable field has the keys.
func (w Wizard) Editing() bool { return w.editing }

// editKey handles a key while the editable field has the keyboard: a printable
// character is added, Backspace removes the last one, Enter or Esc stops.
func (w Wizard) editKey(name string) (Wizard, Outcome) {
	switch name {
	case "enter", "esc":
		w.editing = false
	case "ctrl+c":
		return w, Outcome{Done: true}
	case "backspace":
		if r := []rune(w.text); len(r) > 0 {
			w.text = string(r[:len(r)-1])
		}
	default:
		if r := []rune(name); len(r) == 1 && unicode.IsPrint(r[0]) && len([]rune(w.text)) < maxText {
			w.text += name
		}
	}
	return w, Outcome{}
}

func (w Wizard) askingKey(name string) (Wizard, Outcome) {
	if w.style.Input != "" {
		if w.editing {
			return w.editKey(name)
		}
		if name == "e" {
			w.editing = true
			return w, Outcome{}
		}
	}
	n := len(w.question.Options)
	switch commands.Nav(name) {
	case commands.NavCancel:
		if name == "esc" {
			if b, out, ok := w.goBack(); ok {
				return b, out
			}
		}
		return w, Outcome{Done: true}
	case commands.NavUp:
		w.cursor = (w.cursor - 1 + n) % n
	case commands.NavDown, commands.NavNext:
		w.cursor = (w.cursor + 1) % n
	case commands.NavHome:
		w.cursor = 0
	case commands.NavEnd:
		w.cursor = n - 1
	case commands.NavConfirm:
		return w.choose(w.cursor), Outcome{}
	case commands.NavPageUp, commands.NavPageDown:
		// A long question scrolls by the page, as its "more" hint says.
		return w.scrollKey(name), Outcome{}
	default:
		switch name {
		case "ctrl+u", "ctrl+d":
			return w.scrollKey(name), Outcome{}
		case "left", "b", "backspace":
			if b, out, ok := w.goBack(); ok {
				return b, out
			}
		default:
			if i, err := strconv.Atoi(name); err == nil && i >= 1 && i <= n {
				return w.choose(i - 1), Outcome{}
			}
		}
	}
	return w.follow(), Outcome{}
}

// goBack returns to the screen before: the earlier question, or the stage
// before the dialog (Style.Back). ok is false when no screen before exists.
func (w Wizard) goBack() (Wizard, Outcome, bool) {
	switch {
	case w.canGoBack():
		return w.back(), Outcome{}, true
	case w.style.Back:
		return w, Outcome{Back: true}, true
	}
	return w, Outcome{}, false
}

// choose answers the question with option i.
func (w Wizard) choose(i int) Wizard {
	if w.question.Options[i].Disabled != "" {
		// The option is not an answer: the cursor goes to it, so its reason shows.
		w.cursor = i
		return w.follow()
	}
	w.answers = append(append([]int(nil), w.answers...), i+1)
	return w.resolve()
}

func (w Wizard) reviewKey(name string) (Wizard, Outcome) {
	switch name {
	case "enter", "y":
		return w.begin(), Outcome{Apply: true}
	case "n":
		if w.style.Decline != "" {
			return w, Outcome{Decline: true}
		}
		return w, Outcome{Done: true}
	case "d":
		if w.hasDiff {
			w.full, w.top = !w.full, 0
		}
	case "s":
		if w.style.Skip != "" {
			return w, Outcome{Skip: true}
		}
	case "esc", "left", "b", "backspace":
		if b, out, ok := w.goBack(); ok {
			return b, out
		}
		if name == "esc" {
			return w, Outcome{Done: true}
		}
	case "ctrl+c", "q":
		return w, Outcome{Done: true}
	default:
		w = w.scrollKey(name)
	}
	return w, Outcome{}
}

func (w Wizard) failedKey(name string) (Wizard, Outcome) {
	switch name {
	case "r", "enter":
		if w.style.NoRetry {
			return w, Outcome{}
		}
		return w.begin(), Outcome{Apply: true}
	case "y":
		return w, Outcome{Copy: w.FailureText()}
	case "esc", "ctrl+c", "q", "c":
		return w, Outcome{Done: true}
	}
	return w.scrollKey(name), Outcome{}
}

// scrollKey moves the body of the review or of a failure.
func (w Wizard) scrollKey(name string) Wizard {
	page := pane.PageStep(w.visible())
	action := commands.Nav(name)
	switch name {
	case "ctrl+u":
		action = commands.NavPageUp
	case "ctrl+d":
		action = commands.NavPageDown
	}
	switch action {
	case commands.NavUp:
		return w.scroll(-1)
	case commands.NavDown:
		return w.scroll(1)
	case commands.NavPageUp:
		return w.scroll(-page)
	case commands.NavPageDown:
		return w.scroll(page)
	case commands.NavHome:
		return w.scroll(-w.top)
	case commands.NavEnd:
		return w.scroll(1 << 20)
	}
	return w
}

// scroll moves the body by delta lines, never past its ends.
func (w Wizard) scroll(delta int) Wizard {
	total := len(w.bodyRows())
	if w.LogOpen() {
		total = len(w.log)
	}
	w.top = pane.ClampTop(w.top+delta, total, w.visible())
	if w.LogOpen() {
		w.tailing = w.top >= w.bottom()
	}
	return w
}

// Wheel scrolls the body (a question, the review, a failure's output); it
// never moves the cursor of a question.
func (w Wizard) Wheel(delta int) Wizard {
	if w.phase == Asking || w.phase == Reviewing || w.phase == Failed || w.LogOpen() {
		return w.scroll(delta)
	}
	return w
}

// Hover highlights the button or the option under the pointer.
func (w Wizard) Hover(x, y int) Wizard {
	w.hover = button.At(w.buttonSpans(w.Rect().W-4), w.Rect(), x, y)
	return w
}

// Click handles a left press: a button does what its key does, an option is
// chosen, a press outside the dialog cancels (unless a step is running).
func (w Wizard) Click(x, y int) (Wizard, Outcome) {
	r := w.Rect()
	if !r.Contains(x, y) {
		if w.phase == Applying {
			return w, Outcome{}
		}
		return w, Outcome{Done: true}
	}
	if i := button.At(w.buttonSpans(r.W-4), r, x, y); i >= 0 {
		return w.Key(w.buttons()[i].key)
	}
	if w.phase == Applying && !w.logOpen && w.onLogHint(x, y) {
		return w.Key("l")
	}
	if w.phase == Asking {
		if i, ok := w.optionAt(y); ok {
			return w.choose(i), Outcome{}
		}
	}
	return w, Outcome{}
}
