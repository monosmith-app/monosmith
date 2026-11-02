package releasepick

import (
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func samplePlan() release.Plan {
	return release.Plan{Items: []release.Item{
		{Project: "web", Current: "1.0.0", Next: "1.1.0", Level: "minor", Reason: "2 releasing commit(s)"},
		{Project: "kit", Current: "0.2.0", Next: "0.2.1", Level: "patch", Reason: "dependency web releases"},
		{Project: "sdk", Current: "", Next: "0.1.0", Level: "minor", Reason: "1 releasing commit(s)"},
	}}
}

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		m, _ = m.Key(k)
	}
	return m
}

func screen(m Model) string {
	var rows []string
	for _, l := range m.Lines(theme.Default()) {
		rows = append(rows, ansi.Strip(l))
	}
	return strings.Join(rows, "\n")
}

func TestNothingChosenIsThePlainRelease(t *testing.T) {
	m := New(samplePlan()).Place(100, 30)
	if got := m.Argv(); !slices.Equal(got, []string{"release"}) {
		t.Fatalf("got %v", got)
	}
	if _, out := m.Key("enter"); !out.Done || !out.Confirmed {
		t.Fatalf("Enter goes on: %+v", out)
	}
}

func TestSpaceHoldsAPackageBackAndTheLineSaysSo(t *testing.T) {
	m := press(New(samplePlan()).Place(100, 30), "down", " ")
	if got := m.Argv(); !slices.Equal(got, []string{"release", "--hold", "kit"}) {
		t.Fatalf("got %v", got)
	}
	if !m.Held("kit") || m.Held("web") {
		t.Fatal("kit alone is held")
	}
	if text := screen(m); !strings.Contains(text, "[x] kit") || !strings.Contains(text, "held back: it stays at 0.2.0") || !strings.Contains(text, "$ smith release --hold kit") {
		t.Fatalf("the screen does not show the hold:\n%s", text)
	}
	if got := press(m, " ").Argv(); !slices.Equal(got, []string{"release"}) {
		t.Fatalf("a second Space lets it go: %v", got)
	}
}

func TestHoldingAPackageHoldsWhatReleasesOnlyBecauseOfIt(t *testing.T) {
	m := press(New(samplePlan()).Place(100, 30), " ")
	if text := screen(m); !strings.Contains(text, "a package it depends on is held back") {
		t.Fatalf("kit follows web:\n%s", text)
	}
}

func TestBumpIsPickedPerPackageAndCyclesThroughTheLevels(t *testing.T) {
	m := New(samplePlan()).Place(100, 30)
	for _, want := range []string{"patch", "minor", "major"} {
		m = press(m, "b")
		if got := m.Argv(); !slices.Equal(got, []string{"release", "--bump", "web=" + want}) {
			t.Fatalf("got %v want %s", got, want)
		}
	}
	if got := press(m, "b").Argv(); !slices.Equal(got, []string{"release"}) {
		t.Fatalf("the cycle ends where the commits decide: %v", got)
	}
	if got := press(m, "left").Argv(); !slices.Equal(got, []string{"release", "--bump", "web=minor"}) {
		t.Fatalf("Left goes back: %v", got)
	}
	if text := screen(m);!strings.Contains(text, "1.0.0 → 2.0.0  bump: ‹ major ›") {
		t.Fatalf("the row shows the version the bump gives:\n%s", text)
	}
}

func TestAFirstReleaseHasNoBumpToChoose(t *testing.T) {
	m := press(New(samplePlan()).Place(100, 30), "down", "down", "b")
	if got := m.Argv(); !slices.Equal(got, []string{"release"}) {
		t.Fatalf("got %v", got)
	}
	if text := screen(m); !strings.Contains(text, "no bump: first release") {
		t.Fatalf("it says why:\n%s", text)
	}
}

func TestHoldAndBumpTogetherAndEscCancels(t *testing.T) {
	m := press(New(samplePlan()).Place(100, 30), "b", "b", "b", "down", " ")
	want := []string{"release", "--hold", "kit", "--bump", "web=major"}
	if got := m.Argv(); !slices.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if _, out := m.Key("esc"); !out.Done || out.Confirmed {
		t.Fatalf("Esc cancels: %+v", out)
	}
}

func TestAHeldPackageLosesItsBump(t *testing.T) {
	m := press(New(samplePlan()).Place(100, 30), "b", " ")
	if got := m.Argv(); !slices.Equal(got, []string{"release", "--hold", "web"}) {
		t.Fatalf("got %v", got)
	}
}

func TestAShortScreenScrollsTheListAndKeepsTheCommandAndButtons(t *testing.T) {
	var plan release.Plan
	for _, n := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		plan.Items = append(plan.Items, release.Item{Project: n, Current: "1.0.0", Next: "1.0.1", Level: "patch"})
	}
	m := New(plan).Within(3, 2).Place(80, 16)
	r := m.Rect()
	if r.Y < 3 || r.Y+r.H > 14 {
		t.Fatalf("the dialog (rows %d-%d) leaves its margins", r.Y, r.Y+r.H)
	}
	m = press(m, "down", "down", "down", "down", "down", "down", "down", " ")
	text := screen(m)
	for _, want := range []string{"[x] h", "The command that will run", "$ smith release --hold h", "[ Continue  Enter ]", "packages shown"} {
		if !strings.Contains(text, want) {
			t.Fatalf("lost %q:\n%s", want, text)
		}
	}
}

func TestClickingARowHoldsItAndClickingContinueGoesOn(t *testing.T) {
	m := New(samplePlan()).Place(100, 30)
	r := m.Rect()
	m, _ = m.Click(r.X+3, r.Y+1+m.head+1)
	if !m.Held("kit") {
		t.Fatal("a click on the box of the second row holds kit")
	}
	if _, out := m.Click(r.X+r.W/2-10, r.Y+r.H-2); !out.Done || !out.Confirmed {
		t.Fatalf("a click on Continue goes on: %+v", out)
	}
	if _, out := m.Click(0, 0); !out.Done || out.Confirmed {
		t.Fatalf("a click outside cancels: %+v", out)
	}
}
