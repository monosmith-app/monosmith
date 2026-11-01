package fixwizard

import (
	"strconv"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
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
	review []string
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
			text = "● " + fix.Description + "\n  could not show the changes: " + err.Error()
		}
		w.review = strings.Split(strings.TrimRight(text, "\n"), "\n")
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

// Begin switches to the applying phase, as Outcome.Apply asks the shell to run the fix.
func (w Wizard) begin() Wizard {
	w.phase, w.step, w.total, w.stepName, w.top = Applying, 0, len(w.resolved.Steps), "", 0
	w.headline, w.output = "", nil
	w.started = time.Now()
	return w
}

// Fail shows how the fix stopped and offers Retry and Close.
func (w Wizard) Fail(err error) Wizard {
	w.phase, w.top = Failed, 0
	w.headline, w.output = failureOf(err, w.style.Name)
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
		if name == "ctrl+c" || name == "esc" {
			return w, Outcome{Stop: true}
		}
	case Failed:
		return w.failedKey(name)
	}
	return w, Outcome{}
}

func (w Wizard) askingKey(name string) (Wizard, Outcome) {
	n := len(w.question.Options)
	switch name {
	case "esc", "ctrl+c", "q":
		return w, Outcome{Done: true}
	case "up", "k":
		w.cursor = (w.cursor - 1 + n) % n
	case "down", "j", "tab":
		w.cursor = (w.cursor + 1) % n
	case "home":
		w.cursor = 0
	case "end":
		w.cursor = n - 1
	case "left", "b", "backspace":
		if w.canGoBack() {
			return w.back(), Outcome{}
		}
		if w.style.Back {
			return w, Outcome{Back: true}
		}
	case "enter", " ":
		return w.choose(w.cursor), Outcome{}
	default:
		if i, err := strconv.Atoi(name); err == nil && i >= 1 && i <= n {
			return w.choose(i - 1), Outcome{}
		}
	}
	return w.follow(), Outcome{}
}

// choose answers the question with option i.
func (w Wizard) choose(i int) Wizard {
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
	case "esc", "ctrl+c", "q":
		return w, Outcome{Done: true}
	case "left", "b", "backspace":
		if w.canGoBack() {
			return w.back(), Outcome{}
		}
		if w.style.Back {
			return w, Outcome{Back: true}
		}
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
	page := detail.PageStep(w.visible())
	switch name {
	case "up", "k":
		return w.scroll(-1)
	case "down", "j":
		return w.scroll(1)
	case "pgup", "ctrl+u":
		return w.scroll(-page)
	case "pgdown", "ctrl+d":
		return w.scroll(page)
	case "home":
		return w.scroll(-w.top)
	case "end":
		return w.scroll(1 << 20)
	}
	return w
}

// scroll moves the body by delta lines, never past its ends.
func (w Wizard) scroll(delta int) Wizard {
	w.top = detail.ClampTop(w.top+delta, len(w.bodyRows()), w.visible())
	return w
}

// Wheel scrolls the body (the review, a failure's output).
func (w Wizard) Wheel(delta int) Wizard {
	if w.phase == Reviewing || w.phase == Failed {
		return w.scroll(delta)
	}
	return w
}

// Hover highlights the button or the option under the pointer.
func (w Wizard) Hover(x, y int) Wizard {
	w.hover = -1
	if i := w.buttonAt(x, y); i >= 0 {
		w.hover = i
	}
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
	if i := w.buttonAt(x, y); i >= 0 {
		return w.Key(w.buttons()[i].key)
	}
	if w.phase == Asking {
		if i, ok := w.optionAt(y); ok {
			return w.choose(i), Outcome{}
		}
	}
	return w, Outcome{}
}
