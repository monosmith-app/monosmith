package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/releaseview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runview"
)

// releaseRead replaces what the Release view has read, as the shell would hold it.
func (l *lab) releaseRead(r releaseview.Result) {
	sh := l.sh()
	sh.sources.Release = func() releaseview.Result { return r }
	sh.release = releaseview.FromResult(r)
	l.m = sh
}

// readyToRelease is a plan with nothing to warn about: a clean doctor, a pushed
// branch with a remote and a test run that passed.
func readyToRelease(l *lab) {
	sh := l.sh()
	sh.doctor = sh.doctor.Replace(doctorview.Result{Report: doctor.Report{Passed: 3}})
	test := runview.NewRun([]string{"test"}, nil, l.clock)
	test.End(0, false, l.clock)
	sh.runs = append(sh.runs, test)
	l.m = sh
	r := releaseview.Result{Plan: sh.release.Plan, Git: releaseview.GitState{Known: true, Branch: "main", Remote: "origin", HasRemote: true, Upstream: true}}
	l.releaseRead(r)
}

func firstRelease() releaseview.Result {
	return releaseview.Result{Plan: release.Plan{Unchanged: []string{"lib"}, NeverReleased: []string{"lib"}}}
}

func TestTheReleaseViewHasAReleaseNowButtonThatOpensTheQuestion(t *testing.T) {
	l := newLab(t, true, 120, 34)
	l.keys("6")
	l.has("2 packages to release", "Release now (x)")
	l.clickText("Release now (x)")
	l.has("Release: hold back or bump")
	if len(l.started) != 0 {
		t.Fatal("a click on the button asks first")
	}
}

func TestTheReleaseConfirmationListsTheChecksTheEffectsAndTheCommandAndStartsOnCancel(t *testing.T) {
	l := newLab(t, true, 120, 40)
	l.keys("6", "x", "enter")
	l.has("Release now", "Before you release:", "Doctor: 2 findings are open", "Tests: no test run in this session", "Branch: not known",
		"A warning does not stop the release", "Packages and versions to publish:", "web 1.0.0 → 1.1.0 tag web@1.1.0",
		"What will happen:", "Commit: no", "Tag: 2 tags are created.", "Push: no. There is no remote", "Publish: yes",
		"$ smith release --yes", "Release anyway (y)", "Cancel (esc)")
	l.keys("enter")
	if len(l.started) != 0 {
		t.Fatalf("Enter on the focused Cancel ran %v", l.started)
	}
	l.hasNot("The command that will run:")
}

func TestEscCancelsTheReleaseConfirmation(t *testing.T) {
	l := newLab(t, true, 120, 40)
	l.keys("6", "x", "enter", "esc")
	if len(l.started) != 0 {
		t.Fatalf("Esc ran %v", l.started)
	}
	l.hasNot("The command that will run:")
}

func TestAConfirmationWithNoWarningSaysReleaseNow(t *testing.T) {
	l := newLab(t, true, 120, 40)
	l.keys("6")
	readyToRelease(l)
	l.keys("x", "enter")
	l.has("Doctor: no findings", "Tests: the last test run passed", "Branch main: pushed", "Push: the tags go to origin.", "Release now (y)")
	l.hasNot("Release anyway", "A warning does not stop")
}

func TestTheReleaseRunShowsTheCommandAStepPerPackageAndThenTheResultAndWhatNext(t *testing.T) {
	l := newLab(t, true, 120, 40)
	l.keys("6", "x", "enter")
	p := &scripted{lines: []string{"web: building 1.1.0"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("y")
	if l.sh().tab != "release" {
		t.Fatalf("the release left the Release screen for %q", l.sh().tab)
	}
	l.has("Releasing: smith release --yes", "web  building", "kit  waiting", "Ctrl+C stops it")
	l.send(outputMsg{proc: p, lines: []string{"web: tagged web@1.1.0", "kit: building 0.2.1", "kit: tagged kit@0.2.1"}})
	l.send(outputMsg{proc: p, exit: &runner.Exit{Code: 0}})
	l.has("Released", "New tags:", "web@1.1.0", "kit@0.2.1", "What next:", "Add a remote")
	l.keys("enter")
	l.hasNot("New tags:")
}

func TestAFailedReleaseSaysSoAndWhatToDoNext(t *testing.T) {
	l := newLab(t, true, 120, 40)
	l.keys("6", "x", "enter")
	p := &scripted{lines: []string{"web: building 1.1.0", "web:build x boom"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("y")
	l.send(outputMsg{proc: p, exit: &runner.Exit{Code: 1}})
	l.has("The release stopped", "What next: fix the cause in the output")
	l.hasNot("New tags:")
}

func TestAnEmptyPlanHasOneHeadlineAFoldedWhyAndNoRepeatedText(t *testing.T) {
	l := newLab(t, true, 120, 34)
	l.keys("6")
	l.releaseRead(releaseview.Result{Plan: release.Plan{Unchanged: []string{"sdk"}}})
	l.has("Nothing to release: no feat: or fix: commits since the last tag", "Why? (w)")
	l.hasNot("Only feat:", "Create the first release commit", "Release now")
	l.keys("w")
	l.has("Only feat: (a minor version) and fix: (a patch) commits make a release", "A chore: commit releases nothing")
	l.keys("w")
	l.hasNot("Only feat:")
}

func TestThePreviewOfAnEmptyPlanDoesNotRepeatTheHeadline(t *testing.T) {
	l := newLab(t, true, 120, 34)
	l.keys("6")
	l.releaseRead(releaseview.Result{Plan: release.Plan{Unchanged: []string{"sdk"}}})
	l.next = &scripted{lines: []string{"Dry run: nothing is published or tagged.", "Nothing to release: no package has a feat: or fix: commit since its last release"}}
	l.keys("p")
	l.has("Preview: smith release --dry-run")
	if n := strings.Count(l.view(), "Nothing to release"); n != 1 {
		t.Fatalf("the empty state is said %d times:\n%s", n, l.view())
	}
}

func TestThePreviewOfAPlanLeavesOutTheTerminalAdviceAndPointsAtTheButton(t *testing.T) {
	l := newLab(t, true, 120, 34)
	l.next = &scripted{lines: []string{"web 1.0.0 -> 1.1.0", release.PublishHintLine}}
	l.keys("6", "p")
	l.has("web 1.0.0 -> 1.1.0", "Release now (x) publishes them from here")
	l.hasNot("or run smith release here")
}

func TestFirstReleaseOffersTheCommitAndMakesItOnlyAfterTheConfirmation(t *testing.T) {
	l := newLab(t, true, 120, 34)
	l.keys("6")
	l.releaseRead(firstRelease())
	var made int
	sh := l.sh()
	sh.sources.FirstCommit = func() error { made++; return nil }
	l.m = sh
	l.has("Create the first release commit (c)", `$ git commit --allow-empty -m "feat: first version"`)
	l.keys("c")
	l.has("Create the first release commit", "Cancel (esc)")
	l.keys("enter")
	if made != 0 {
		t.Fatal("Enter on the focused Cancel made the commit")
	}
	l.keys("c", "y")
	if made != 1 {
		t.Fatalf("the confirmed commit was made %d times", made)
	}
}

func TestAFailedFirstCommitIsToldInADialog(t *testing.T) {
	l := newLab(t, true, 120, 34)
	l.keys("6")
	l.releaseRead(firstRelease())
	sh := l.sh()
	sh.sources.FirstCommit = func() error { return errors.New("hook said no") }
	l.m = sh
	l.keys("c", "y")
	l.has("The commit failed: hook said no")
}

func TestTheFirstCommitKeyDoesNotShadowTypecheckWhenThereIsAPlan(t *testing.T) {
	l := newLab(t, true, 120, 34)
	l.keys("6", "c")
	if got := l.lastStarted(); !strings.Contains(got, "typecheck") {
		t.Fatalf("c started %q, want the typecheck run", got)
	}
}
