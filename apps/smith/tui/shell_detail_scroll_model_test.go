package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/splitter"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// longProblem is what a check that found a problem in every project says: one
// sentence per project, joined with "; ".
func longProblem(n int) string {
	var parts []string
	for i := 1; i <= n; i++ {
		parts = append(parts, fmt.Sprintf("pkg%02d:build writes packages/pkg%02d/dist but its task caches dist/packages/pkg%02d: a cache hit would leave dist empty", i, i, i))
	}
	return strings.Join(parts, "; ")
}

// doctorLab is a wide Doctor tab whose first check has a problem of 23 sentences.
func doctorLab(t *testing.T, w, h int) *lab {
	t.Helper()
	l := newLab(t, true, w, h)
	report := doctor.Report{Failed: 1, Findings: []doctor.Finding{
		{Finding: protocol.Finding{Check: "every build declares its outputs", Detail: longProblem(23), Remedy: "declare dist as the output of each build, then run smith doctor again"}, Source: "node"},
		{Finding: protocol.Finding{Check: "second check", Detail: "short", Remedy: "do this"}, Source: "go"},
	}}
	l.send(doctorview.Result{Report: report})
	l.keys("4")
	return l
}

func TestALongProblemFoldsSoTheRemedyAndTheActionsStayOnTheScreen(t *testing.T) {
	l := doctorLab(t, 140, 30)
	l.has("How to fix", "declare dist as the output", "Actions", "[Check again (r)]", "What is wrong", "pkg01:build", "23, e shows all")
	l.hasNot("pkg06:build")
	v := l.view()
	if a, w, h := strings.Index(v, "Actions"), strings.Index(v, "What is wrong"), strings.Index(v, "How to fix"); a >= w || w >= h {
		t.Fatalf("the order is state, actions, what is wrong, how to fix:\n%s", v)
	}
}

// TestEnterOnThePassedLineShowsThePassedChecksBesideTheDetail keeps the promise
// of the line "enter or p shows them" when the detail sits beside the list.
func TestEnterOnThePassedLineShowsThePassedChecksBesideTheDetail(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.send(doctorview.Result{Report: doctor.Report{Passed: 2, Findings: []doctor.Finding{
		{Finding: protocol.Finding{Check: "git is installed", OK: true}},
		{Finding: protocol.Finding{Check: "smith.toml reads", OK: true}},
	}}})
	l.keys("4")
	l.has("2 checks passed")
	l.hasNot("git is installed")
	l.keys("enter")
	l.has("git is installed", "smith.toml reads")
}

func TestEnterAndEUnfoldTheProblemAndTheDetailScrolls(t *testing.T) {
	for _, key := range []string{"enter", "e"} {
		l := doctorLab(t, 140, 30)
		l.keys(key)
		l.hasNot("e shows all")
		l.has("↓ more", "pkg01:build")
		l.hasNot("pkg23:build")
		for i := 0; i < 6; i++ {
			l.keys("ctrl+d")
		}
		l.has("pkg23:build", "↑ more")
		l.hasNot("↓ more", "pkg01:build")
		l.keys("ctrl+u", "ctrl+u", "ctrl+u", "ctrl+u", "ctrl+u", "ctrl+u")
		l.has("pkg01:build")
		l.keys(key)
		l.has("23, e shows all")
	}
}

func TestPageKeysAndTheWheelScrollTheDetail(t *testing.T) {
	l := doctorLab(t, 140, 30)
	l.keys("e", "pgdown")
	l.has("↑ more")
	l.keys("pgup")
	l.hasNot("↑ more")
	x, y := l.find("What is wrong")
	l.send(wheel(x, y, 1))
	l.has("↑ more")
	l.send(wheel(x, y, -1))
	l.hasNot("↑ more")
}

func TestAnotherCheckStartsAtTheTopAndFolded(t *testing.T) {
	l := doctorLab(t, 140, 30)
	l.keys("e", "pgdown", "down")
	l.has("second check", "do this")
	l.hasNot("↑ more")
	l.keys("up")
	l.has("23, e shows all")
}

func TestBracketsMoveTheDividerAndTheWidthIsKeptForThatView(t *testing.T) {
	l := doctorLab(t, 140, 30)
	before := l.sh().lay().Detail.W
	l.keys("]")
	if got := l.sh().lay().Detail.W; got != before-splitter.Step {
		t.Fatalf("] moves the divider right, so the detail narrows by %d: %d -> %d", splitter.Step, before, got)
	}
	l.keys("[", "[")
	if got := l.sh().lay().Detail.W; got != before+splitter.Step {
		t.Fatalf("[ moves the divider left: %d", got)
	}
	l.keys("2")
	if got := l.sh().lay().Detail.W; got == before+splitter.Step {
		t.Fatalf("each view keeps its own width, Projects has %d", got)
	}
	l.keys("4")
	if got := l.sh().lay().Detail.W; got != before+splitter.Step {
		t.Fatalf("Doctor keeps the width it was given: %d", got)
	}
	for i := 0; i < 40; i++ {
		l.keys("[")
	}
	if got := l.sh().lay().Main.W; got < 36 {
		t.Fatalf("the list keeps its room: %d", got)
	}
}

func TestZoomShowsTheDetailAloneAndGivesTheListBack(t *testing.T) {
	l := doctorLab(t, 140, 30)
	l.has("second check")
	l.keys("z")
	l.hasNot("second check")
	l.has("How to fix", "Check again")
	l.keys("e", "pgdown")
	l.has("↑ more")
	l.keys("z")
	l.has("second check")
}

func TestOnANarrowScreenZoomSwitchesBetweenTheListAndTheDetail(t *testing.T) {
	l := doctorLab(t, 90, 30)
	l.has("second check")
	l.hasNot("How to fix")
	l.keys("z")
	l.has("How to fix", "Check again", "23, e shows all")
	l.hasNot("second check")
	l.keys("pgdown", "esc")
	l.has("second check")
	l.hasNot("How to fix")
	l.keys("z", "z")
	l.has("second check")
}

func TestZoomWorksInTheProjectsViewToo(t *testing.T) {
	l := newLab(t, true, 140, 30)
	l.keys("z")
	l.has("Run on orders")
	l.keys("z")
	l.has("Projects (")
}

func TestTheDetailKeysAreInTheHelp(t *testing.T) {
	l := doctorLab(t, 140, 30)
	l.keys("?")
	l.has("Scroll the detail up", "Move the selected divider right or down", "Zoom the detail")
}
