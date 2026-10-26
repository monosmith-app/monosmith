package fixwizard

import (
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func plain() theme.Theme { return theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode}) }

// show is a fake engine: it names the files and steps of the resolved fix.
func show(p doctorfix.Planned) (string, error) {
	var b strings.Builder
	b.WriteString("● " + p.Description + "\n")
	for _, f := range p.Files {
		b.WriteString("--- " + f.Path + "\n+++ " + f.Path + "\n@@ -1 +1 @@\n-old\n+" + strings.TrimSpace(f.Content) + "\n")
	}
	for i, s := range p.Steps {
		b.WriteString("  step " + string(rune('1'+i)) + ": $ " + strings.Join(s.Argv, " ") + "\n")
	}
	return b.String(), nil
}

func question() doctorfix.Planned {
	pm := func(name string) protocol.FixChoiceOption {
		return protocol.FixChoiceOption{Label: name, Fix: protocol.Fix{
			Files: []protocol.FileEdit{{Path: "package.json", Content: name}},
			Steps: []protocol.FixStep{{Argv: []string{name, "install"}}},
		}}
	}
	return doctorfix.Planned{
		Fix: protocol.Fix{Description: "install the dependencies", Choices: []protocol.FixChoice{
			{Question: "Which package manager?", Options: []protocol.FixChoiceOption{pm("npm"), pm("pnpm")}},
			{Question: "Install now?", Options: []protocol.FixChoiceOption{
				{Label: "yes", Fix: protocol.Fix{Steps: []protocol.FixStep{{Argv: []string{"echo", "now"}}}}},
				{Label: "later"},
			}},
		}},
		Checks: []string{"deps installed"}, Ask: true,
	}
}

func safe() doctorfix.Planned {
	return doctorfix.Planned{Fix: protocol.Fix{
		Description: "ignore the cache", Safe: true, Files: []protocol.FileEdit{{Path: ".gitignore", Content: ".smith/"}},
	}, Checks: []string{"cache ignored"}}
}

func press(t *testing.T, w Wizard, keys ...string) (Wizard, Outcome) {
	t.Helper()
	var out Outcome
	for _, k := range keys {
		w, out = w.Key(k)
	}
	return w, out
}

func text(w Wizard) string {
	w = w.Place(100, 30)
	return pane.Strip(strings.Join(w.Lines(plain(), 0), "\n"))
}

func TestAFixWithQuestionsAsksThemInOrderThenReviewsTheResolvedFix(t *testing.T) {
	w := New(question(), show).Place(100, 30)
	if w.Phase() != Asking {
		t.Fatalf("phase %v, want Asking", w.Phase())
	}
	if got := text(w); !strings.Contains(got, "question 1 of 2") || !strings.Contains(got, "Which package manager?") || !strings.Contains(got, "1) npm  (suggested)") || !strings.Contains(got, "2) pnpm") {
		t.Fatalf("first question:\n%s", got)
	}
	w, _ = press(t, w, "down", "enter") // pnpm
	if got := text(w); !strings.Contains(got, "question 2 of 2") || !strings.Contains(got, "Install now?") {
		t.Fatalf("second question:\n%s", got)
	}
	w, out := press(t, w, "1") // a number picks at once
	if out != (Outcome{}) || w.Phase() != Reviewing {
		t.Fatalf("after the last answer: phase %v, outcome %+v", w.Phase(), out)
	}
	got := text(w)
	for _, want := range []string{"Fix · review", "package.json", "+pnpm", "pnpm install", "echo now", "[ Apply  Enter ]", "[ Back  b ]"} {
		if !strings.Contains(got, want) {
			t.Fatalf("review lacks %q:\n%s", want, got)
		}
	}
	fix := w.Resolved()
	if len(fix.Steps) != 2 || fix.Steps[0].Argv[0] != "pnpm" || len(fix.Choices) != 0 {
		t.Fatalf("resolved %+v", fix)
	}
}

func TestTheArrowsWrapAndJAndKMoveLikeThem(t *testing.T) {
	w := New(question(), show).Place(100, 30)
	w, _ = press(t, w, "up")
	if w.cursor != 1 {
		t.Fatalf("up from the first option went to %d", w.cursor)
	}
	w, _ = press(t, w, "j", "k", "k")
	if w.cursor != 0 {
		t.Fatalf("j k k left the cursor on %d", w.cursor)
	}
}

func TestBackUndoesTheLastAnswerAndKeepsTheCursorOnIt(t *testing.T) {
	w := New(question(), show).Place(100, 30)
	w, _ = press(t, w, "2", "b")
	if w.Phase() != Asking || len(w.answers) != 0 || w.cursor != 1 {
		t.Fatalf("back: phase %v answers %v cursor %d", w.Phase(), w.answers, w.cursor)
	}
	w, _ = press(t, w, "left") // nothing to go back to: stays
	if w.Phase() != Asking || len(w.answers) != 0 {
		t.Fatal("left on the first question moved")
	}
}

func TestEscCancelsAtEveryPhaseThatAsksAndNumbersOutOfRangeDoNothing(t *testing.T) {
	w := New(question(), show).Place(100, 30)
	if _, out := press(t, w, "9"); out.Done || len(w.answers) != 0 {
		t.Fatal("option 9 of 2 did something")
	}
	if _, out := press(t, w, "esc"); !out.Done || out.Apply {
		t.Fatalf("esc while asking: %+v", out)
	}
	r := New(safe(), show).Place(100, 30)
	if _, out := press(t, r, "esc"); !out.Done || out.Apply {
		t.Fatalf("esc in the review: %+v", out)
	}
}

func TestAFixWithoutQuestionsStartsOnTheReviewAndNeedsOneConfirmation(t *testing.T) {
	w := New(safe(), show).Place(100, 30)
	if w.Phase() != Reviewing {
		t.Fatalf("phase %v, want Reviewing", w.Phase())
	}
	if got := text(w); !strings.Contains(got, ".gitignore") || strings.Contains(got, "[ Back") {
		t.Fatalf("review:\n%s", got)
	}
	w, out := press(t, w, "enter")
	if !out.Apply || w.Phase() != Applying {
		t.Fatalf("enter: phase %v outcome %+v", w.Phase(), out)
	}
}

func TestApplyingShowsTheStepAndOnlyCtrlCStopsIt(t *testing.T) {
	w := New(question(), show).Place(100, 30)
	w, _ = press(t, w, "1", "1", "enter")
	w = w.Progress(2, 2, "echo now")
	if got := text(w); !strings.Contains(got, "step 2 of 2: echo now") || !strings.Contains(got, "[ Stop  Ctrl+C ]") {
		t.Fatalf("applying:\n%s", got)
	}
	if w2, out := press(t, w, "enter", "q", "x"); out != (Outcome{}) || w2.Phase() != Applying {
		t.Fatalf("a key other than stop did something: %+v", out)
	}
	if _, out := press(t, w, "ctrl+c"); !out.Stop {
		t.Fatalf("ctrl+c: %+v", out)
	}
}

func TestAStepErrorShowsTheStepAndItsOutputAndOffersRetryAndClose(t *testing.T) {
	w := New(safe(), show).Place(100, 30)
	w, _ = press(t, w, "enter")
	w = w.Fail(&doctorfix.StepError{Number: 2, Total: 3, Step: protocol.FixStep{Argv: []string{"npm", "install"}}, Err: errors.New("exited with code 1"), Output: "npm ERR! 404 not found\nnpm ERR! left-pad"})
	got := text(w)
	for _, want := range []string{"Fix failed", "Step 2 of 3 failed: npm install", "exited with code 1", "npm ERR! 404 not found", "[ Retry  r ]", "[ Copy output  y ]", "[ Close  Esc ]"} {
		if !strings.Contains(got, want) {
			t.Fatalf("failure lacks %q:\n%s", want, got)
		}
	}
	if _, out := press(t, w, "y"); !strings.Contains(out.Copy, "npm ERR! left-pad") || !strings.Contains(out.Copy, "Step 2 of 3") {
		t.Fatalf("y copied %q", out.Copy)
	}
	r, out := press(t, w, "r")
	if !out.Apply || r.Phase() != Applying {
		t.Fatalf("retry: phase %v outcome %+v", r.Phase(), out)
	}
	if _, out := press(t, w, "esc"); !out.Done {
		t.Fatalf("close: %+v", out)
	}
}

func TestAFailureThatIsNotAStepIsToldInOneLine(t *testing.T) {
	w := New(safe(), show).Place(100, 30)
	w, _ = press(t, w, "enter")
	w = w.Fail(errors.New("a fix may only change files in the workspace, not ../x"))
	if got := text(w); !strings.Contains(got, "Could not apply the fix: a fix may only change files") || strings.Contains(got, "What it printed") {
		t.Fatalf("failure:\n%s", got)
	}
}

func TestALongReviewScrollsWithTheKeysAndTheWheel(t *testing.T) {
	long := doctorfix.Planned{Fix: protocol.Fix{Description: "big", Safe: true}}
	many := func(doctorfix.Planned) (string, error) {
		var b strings.Builder
		for i := 1; i <= 80; i++ {
			b.WriteString("+line " + strings.Repeat("x", i%3) + string(rune('a'+i%26)) + "\n")
		}
		return b.String(), nil
	}
	w := New(long, many).Place(100, 20)
	if got := text2(w, 100, 20); !strings.Contains(got, "↓ more") || strings.Contains(got, "↑ more") {
		t.Fatalf("top of the review:\n%s", got)
	}
	w, _ = press(t, w, "pgdown")
	if got := text2(w, 100, 20); !strings.Contains(got, "↑ more") || !strings.Contains(got, "↓ more") {
		t.Fatalf("after pgdown:\n%s", got)
	}
	w, _ = press(t, w, "end")
	if got := text2(w, 100, 20); strings.Contains(got, "↓ more") || !strings.Contains(got, "↑ more") {
		t.Fatalf("at the end:\n%s", got)
	}
	w = w.Wheel(-1 << 20)
	if w.top != 0 {
		t.Fatalf("wheel up left top at %d", w.top)
	}
	if _, out := press(t, w, "enter"); !out.Apply {
		t.Fatal("a scrolled review could not be applied")
	}
}

func text2(w Wizard, sw, sh int) string {
	w = w.Place(sw, sh)
	lines := w.Lines(plain(), 0)
	return pane.Strip(strings.Join(lines, "\n"))
}

func TestClicksChooseOptionsPressButtonsAndCancelOutside(t *testing.T) {
	w := New(question(), show).Place(100, 30)
	r := w.Rect()
	// The second option's row: find it in the drawn lines.
	var optionY int
	for y, l := range w.Lines(plain(), 0) {
		if strings.Contains(pane.Strip(l), "2) pnpm") {
			optionY = r.Y + y
		}
	}
	w2, out := w.Click(r.X+5, optionY)
	if out != (Outcome{}) || len(w2.answers) != 1 || w2.answers[0] != 2 {
		t.Fatalf("clicking option 2: answers %v outcome %+v", w2.answers, out)
	}
	if _, out := w.Click(0, 0); !out.Done {
		t.Fatal("a click outside did not cancel")
	}
	// Cancel button: the last button of the bottom row.
	spans := w.buttonSpans(r.W - 4)
	last := spans[len(spans)-1]
	if _, out := w.Click(r.X+last[0]+1, r.Y+r.H-2); !out.Done {
		t.Fatalf("the Cancel button: %+v", out)
	}
	rev := New(safe(), show).Place(100, 30)
	rr := rev.Rect()
	if got, out := rev.Click(rr.X+rev.buttonSpans(rr.W - 4)[0][0]+1, rr.Y+rr.H-2); !out.Apply || got.Phase() != Applying {
		t.Fatalf("the Apply button: %+v", out)
	}
	// A press outside while a step runs does not cancel it.
	run, _ := rev.Key("enter")
	if _, out := run.Click(0, 0); out != (Outcome{}) {
		t.Fatalf("a click outside during a step: %+v", out)
	}
}

func TestTheDialogFitsAScreenAtTheMinimumSize(t *testing.T) {
	w := New(question(), show).Place(72, 9)
	lines := w.Lines(plain(), 0)
	r := w.Rect()
	if len(lines) != r.H || r.H > 9 || r.X+r.W > 72 {
		t.Fatalf("rect %+v, %d lines", r, len(lines))
	}
	for _, l := range lines {
		if pane.Width(l) != r.W {
			t.Fatalf("a line is %d cells, want %d: %q", pane.Width(l), r.W, pane.Strip(l))
		}
	}
	// The option under the cursor stays on screen as it moves.
	w, _ = press(t, w, "down")
	if got := pane.Strip(strings.Join(w.Lines(plain(), 0), "\n")); !strings.Contains(got, "2) pnpm") {
		t.Fatalf("cursor option hidden:\n%s", got)
	}
}

func TestAStyledReviewCanDeclineSkipAndGoBackToTheStageBefore(t *testing.T) {
	st := Style{Name: "Adopt", Apply: "Yes, delete", Decline: "No, keep", Skip: "Skip rest", Back: true}
	w := New(safe(), show).Styled(st).Place(100, 30)
	got := text(w)
	for _, want := range []string{"Adopt · review", "[ Yes, delete  Enter ]", "[ No, keep  n ]", "[ Skip rest  s ]", "[ Back  b ]", "[ Cancel  Esc ]"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q:\n%s", want, got)
		}
	}
	for key, check := range map[string]func(Outcome) bool{
		"n": func(o Outcome) bool { return o.Decline }, "s": func(o Outcome) bool { return o.Skip },
		"b": func(o Outcome) bool { return o.Back }, "esc": func(o Outcome) bool { return o.Done }, "enter": func(o Outcome) bool { return o.Apply },
	} {
		if _, out := w.Key(key); !check(out) {
			t.Errorf("%s: %+v", key, out)
		}
	}
	// Without the styles n cancels and s does nothing: the fix wizard is unchanged.
	plainWizard := New(safe(), show).Place(100, 30)
	if _, out := plainWizard.Key("n"); !out.Done || out.Decline {
		t.Errorf("n in the fix wizard: %+v", out)
	}
	if _, out := plainWizard.Key("s"); out != (Outcome{}) {
		t.Errorf("s in the fix wizard: %+v", out)
	}
}

func TestAStartedStepShowsItsHeadlineAndAFailureWithoutRetryOffersOnlyClose(t *testing.T) {
	w := New(safe(), show).Styled(Style{Name: "Adopt", Verb: "Converting…", Working: "reading the files", NoRetry: true}).Place(100, 30).Start("Converting the workspace", 2)
	if got := text(w); !strings.Contains(got, "Converting…") || !strings.Contains(got, "Converting the workspace") || !strings.Contains(got, "reading the files, then 2 step(s)") {
		t.Fatalf("started:\n%s", got)
	}
	w = w.Progress(1, 2, "convert")
	if got := text(w); !strings.Contains(got, "step 1 of 2: convert") {
		t.Fatalf("progress:\n%s", got)
	}
	w = w.Fail(errors.New("boom"))
	got := text(w)
	if strings.Contains(got, "Retry") || !strings.Contains(got, "Adopt failed") || !strings.Contains(got, "[ Close  Esc ]") {
		t.Fatalf("failure:\n%s", got)
	}
	if _, out := w.Key("r"); out.Apply {
		t.Fatal("r retried a step that cannot be retried")
	}
}

func TestAJobWithoutAStepCountShowsTheNameOfWhatItIsDoing(t *testing.T) {
	w := New(safe(), show).Place(100, 30).Start("Reading the repository…", 0)
	w = w.Progress(1, 0, "reading 1,240 files")
	got := text(w)
	if !strings.Contains(got, "Reading 1,240 files…") || strings.Contains(got, "step 1 of") {
		t.Fatalf("progress without a total:\n%s", got)
	}
}

func TestALongOptionWrapsUnderItsNumberAndAClickOnAnyOfItsLinesChoosesIt(t *testing.T) {
	long := strings.Repeat("a very long label ", 8)
	p := doctorfix.Planned{Fix: protocol.Fix{Description: "d", Choices: []protocol.FixChoice{{Question: "Which?", Options: []protocol.FixChoiceOption{{Label: "short"}, {Label: long}}}}}}
	w := New(p, show).Place(60, 30)
	got := text(w)
	if !strings.Contains(got, "2) a very long label") || strings.Count(got, "label") < 8 {
		t.Fatalf("the long option is cut:\n%s", got)
	}
	r := w.Rect()
	_, at := w.rows(plain(), 0)
	if at[2]-at[1] < 2 {
		t.Fatalf("the long option did not wrap: %v", at)
	}
	w2, _ := w.Click(r.X+5, r.Y+1+at[1]+1) // the second line of option 2
	if w2.Phase() != Reviewing || w2.Answers()[0] != 2 {
		t.Fatalf("a click on the wrapped line chose %v in phase %v", w2.Answers(), w2.Phase())
	}
}
