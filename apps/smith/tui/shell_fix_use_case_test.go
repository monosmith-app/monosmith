package tui

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/fixwizard"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// withInstallFix replaces the lab's doctor report with one whose check has a
// fix that asks which package manager to use and installs with it.
func withInstallFix(l *lab) *lab {
	opt := func(name string) protocol.FixChoiceOption {
		return protocol.FixChoiceOption{Label: name, Fix: protocol.Fix{Steps: []protocol.FixStep{{Description: "install with " + name, Argv: []string{name, "install"}}}}}
	}
	report := doctor.Report{Failed: 2, Findings: []doctor.Finding{
		{Finding: protocol.Finding{Check: "dependencies are installed", Detail: "node_modules is missing", Remedy: "install them", Fix: &protocol.Fix{
			Description: "install the dependencies", Choices: []protocol.FixChoice{{Question: "Which package manager?", Options: []protocol.FixChoiceOption{opt("npm"), opt("pnpm")}}}}}},
		{Finding: protocol.Finding{Check: "git hooks are set", Detail: "none", Remedy: "run the hook installer"}},
	}}
	sh := l.sh()
	sh.sources.Doctor = func() doctorview.Result { return doctorview.Result{Report: report} }
	l.m = sh
	return l.keys("4", "r")
}

func TestFixWizardAsksTheQuestionsThenReviewsAndAppliesOffTheUIGoroutine(t *testing.T) {
	l := withInstallFix(newLab(t, true, 140, 34))
	l.has("Fix: install the dependencies", "asks: Which package manager?")
	l.keys("f")
	l.has("Fix · question 1 of 1", "Which package manager?", "1) npm  (suggested)", "2) pnpm")
	l.keys("down", "enter")
	l.has("Fix · review", "$ pnpm install", "[ Apply  Enter ]")
	if len(l.appliedFixes()) != 0 {
		t.Fatal("applied before the confirmation")
	}
	l.keys("enter")
	applied := l.appliedFixes()
	if len(applied) != 1 || len(applied[0].Steps) != 1 || applied[0].Steps[0].Argv[0] != "pnpm" || len(applied[0].Choices) != 0 {
		t.Fatalf("applied %+v", applied)
	}
	if l.sh().fix != nil || l.sh().fixApplying() {
		t.Fatal("the wizard stays open after a success")
	}
	l.has("Fixed: install the dependencies")
}

func TestFixWizardShowsProgressWhileAStepRunsAndCtrlCStopsIt(t *testing.T) {
	l := withInstallFix(newLab(t, true, 140, 34))
	l.applyHold = make(chan struct{})
	l.keys("f", "enter", "enter") // npm, then apply
	l.has("Fixing…", "step 1 of 1: npm install", "Ctrl+C stops")
	if !l.sh().fixApplying() || l.sh().fix.Phase() != fixwizard.Applying {
		t.Fatal("the fix is not applying")
	}
	l.keys("q", "esc") // esc stops, q is ignored: only the first press of a stop key matters
	cancelled := false
	for i := 0; i < 200 && !cancelled; i++ { // the step's goroutine notices a moment later
		l.fixMu.Lock()
		cancelled = l.applyCancel
		l.fixMu.Unlock()
		if !cancelled {
			time.Sleep(10 * time.Millisecond)
		}
	}
	if !cancelled {
		t.Fatal("the step's context was not cancelled")
	}
	run := l.sh().fixing
	l.send(fixDoneMsg{run: run, err: &doctorfix.StepError{Number: 1, Total: 1, Step: protocol.FixStep{Argv: []string{"npm", "install"}}, Err: errors.New("context canceled")}})
	l.has("Fix failed", "Step 1 of 1 failed: npm install", "[ Retry  r ]")
	close(l.applyHold)
}

func TestFixWizardShowsTheFailingStepsOutputAndRetryRunsTheFixAgain(t *testing.T) {
	l := withInstallFix(newLab(t, true, 140, 34))
	sh := l.sh()
	sh.clip = &bytes.Buffer{}
	l.m = sh
	l.applyErr = &doctorfix.StepError{Number: 1, Total: 1, Step: protocol.FixStep{Argv: []string{"npm", "install"}}, Err: errors.New("exited with code 1"), Output: "npm ERR! 404 not found"}
	l.keys("f", "enter", "enter")
	l.has("Fix failed", "Step 1 of 1 failed: npm install", "exited with code 1", "npm ERR! 404 not found", "[ Retry  r ]", "[ Close  Esc ]")
	if len(l.appliedFixes()) != 1 {
		t.Fatalf("applied %d times", len(l.appliedFixes()))
	}
	l.keys("y")
	l.has("copied: Step 1 of 1 failed")
	l.applyErr = nil
	l.keys("r")
	if len(l.appliedFixes()) != 2 || l.sh().fix != nil {
		t.Fatalf("retry applied %d times, wizard open %v", len(l.appliedFixes()), l.sh().fix != nil)
	}
	l.has("Fixed: install the dependencies")
}

func TestFixWizardCancelsFromTheQuestionsAndTheReviewWithoutChangingAnything(t *testing.T) {
	l := withInstallFix(newLab(t, true, 140, 34))
	l.keys("f", "esc")
	l.hasNot("Fix · question")
	l.has("nothing was changed")
	l.keys("f", "1", "esc")
	if l.sh().fix != nil || len(l.appliedFixes()) != 0 {
		t.Fatal("cancelling applied something")
	}
	l.keys("f", "1", "b")
	l.has("Fix · question 1 of 1")
}

func TestFixWizardFixAllReviewsTheSafeFixesOnceAndAppliesThemOnConfirm(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("4", "x")
	l.has("Fix · review", "write web's registry", "apps/web/Dockerfile", "[ Apply  Enter ]")
	if len(l.started) != 0 || len(l.appliedFixes()) != 0 {
		t.Fatal("x applied before the review was confirmed")
	}
	l.keys("esc")
	l.has("nothing was changed")
	l.keys("x", "enter")
	if got := l.appliedFixes(); len(got) != 1 || got[0].Description != "write web's registry" {
		t.Fatalf("applied %+v", got)
	}
	if len(l.started) != 0 {
		t.Fatalf("the fix ran as a command: %v", l.started)
	}
}

func TestFixWizardIsInThePaletteAndRefusesWhenThereIsNothingToFix(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("ctrl+p").typeText("Fix with doctor").keys("enter")
	l.has("Fix · review")
	l = newLab(t, true, 140, 34)
	sh := l.sh()
	sh.sources.Doctor = func() doctorview.Result {
		return doctorview.Result{Report: doctor.Report{Passed: 1, Findings: []doctor.Finding{{Finding: protocol.Finding{Check: "git", OK: true}}}}}
	}
	l.m = sh
	l.keys("4", "r", "x")
	if l.sh().fix != nil {
		t.Fatal("a wizard opened with nothing to fix")
	}
	l.has("nothing it can fix")
}

func TestFixWizardWorksWithTheMouse(t *testing.T) {
	l := withInstallFix(newLab(t, true, 140, 34))
	l.clickText("Fix: install the dependencies")
	l.has("Which package manager?")
	l.clickText("2) pnpm")
	l.has("Fix · review", "$ pnpm install")
	l.clickText("[ Apply")
	if got := l.appliedFixes(); len(got) != 1 || got[0].Steps[0].Argv[0] != "pnpm" {
		t.Fatalf("applied %+v", got)
	}
	// A click outside the dialog cancels it.
	l.keys("f")
	l.send(click(0, 0))
	if l.sh().fix != nil {
		t.Fatal("a click outside kept the wizard open")
	}
}

func TestFixWizardReviewScrollsWithTheWheel(t *testing.T) {
	l := newLab(t, true, 100, 24)
	var content strings.Builder
	for i := 0; i < 60; i++ {
		content.WriteString("line " + strings.Repeat("x", i%7) + "\n")
	}
	sh := l.sh()
	sh.sources.Doctor = func() doctorview.Result {
		return doctorview.Result{Report: doctor.Report{Failed: 1, Findings: []doctor.Finding{{Finding: protocol.Finding{Check: "big file", Fix: &protocol.Fix{
			Description: "write a big file", Safe: true, Files: []protocol.FileEdit{{Path: "big.txt", Content: content.String()}}}}}}}}
	}
	l.m = sh
	l.keys("4", "r", "f")
	l.has("Fix · review", "↓ more")
	l.hasNot("↑ more")
	l.send(wheel(50, 10, 1))
	l.has("↑ more")
	l.keys("pgdown", "end")
	l.hasNot("↓ more")
}

func TestFixWizardDoctorKeepsItsPlaceAfterTheRecheck(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("4", "down")
	before := l.sh().doctor.Cursor
	l.keys("f", "enter")
	if l.sh().doctor.Cursor != before || l.sh().doctor.Report == nil {
		t.Fatalf("cursor %d, was %d", l.sh().doctor.Cursor, before)
	}
}
