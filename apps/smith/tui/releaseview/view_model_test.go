package releaseview

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func planned() Model {
	return FromResult(Result{Plan: release.Plan{
		Items: []release.Item{
			{Project: "web", Plugin: "node", Current: "1.0.0", Next: "2.0.0", Level: "major", Tag: "web@2.0.0", Reason: "3 releasing commit(s)", Changes: []release.Change{
				{Type: "feat", Scopes: []string{"ui", "theme"}, Description: "dark mode", Hash: "a1b2c3d4e5f6", Author: "Ada"},
				{Type: "fix", Description: "cookie parsing", Breaking: true, BreakingNote: "drops the old format", Hash: "9f8e7d6"},
			}},
			{Project: "kit", Plugin: "go", Current: "0.2.0", Next: "0.2.1", Level: "patch", Tag: "kit@0.2.1", Reason: "dependency web releases"},
		},
		Unchanged: []string{"sdk"},
	}})
}

func render(m Model) string {
	lines, _ := m.View(theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII}), 120)
	return strings.Join(lines, "\n")
}

func press(m Model, name string) Model {
	switch name {
	case "down":
		return m.Update(tea.KeyMsg{Type: tea.KeyDown})
	case "up":
		return m.Update(tea.KeyMsg{Type: tea.KeyUp})
	}
	return m.Update(tea.KeyMsg{Type: tea.KeyEnter})
}

func TestEachProjectIsAPreviewLineAndEnterShowsTheCommitsBehindItsBump(t *testing.T) {
	m := planned()
	out := render(m)
	for _, want := range []string{"web", "1.0.0", "2.0.0", "major", "kit", "0.2.1", "patch", "unchanged: sdk"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "a1b2c3d") {
		t.Fatalf("the commits show before the drill-down:\n%s", out)
	}
	out = render(press(m, "enter"))
	for _, want := range []string{"tag web@2.0.0", "a1b2c3d feat(ui,theme): dark mode", "Ada", "9f8e7d6 fix!: cookie parsing", "BREAKING: drops the old format"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "a1b2c3d4e5f6") {
		t.Fatalf("the hash is shortened like git does:\n%s", out)
	}
}

func TestADependencyBumpSaysNoCommitOfItsOwnReleasesIt(t *testing.T) {
	out := render(press(press(planned(), "down"), "enter"))
	if !strings.Contains(out, "dependency web releases (no commit of its own releases it)") {
		t.Fatalf("%s", out)
	}
}

func TestMovingClosesTheDrillDownAndWrapsAround(t *testing.T) {
	m := press(planned(), "enter")
	m = press(m, "down")
	if m.Open || m.Cursor != 1 {
		t.Fatalf("cursor %d open %v", m.Cursor, m.Open)
	}
	if m = press(m, "down"); m.Cursor != 0 {
		t.Fatalf("down from the last wraps to the first: %d", m.Cursor)
	}
	if m = m.SelectRow(1); m.Cursor != 1 || m.Open {
		t.Fatalf("a first click selects: %d %v", m.Cursor, m.Open)
	}
	if m = m.SelectRow(1); !m.Open {
		t.Fatal("a second click opens")
	}
}

func TestOnlyAPlanWithoutBlockersCanBeReleasedOrDrilledInto(t *testing.T) {
	if !planned().CanRelease() {
		t.Fatal("a plan with items can be released")
	}
	blocked := FromResult(Result{Plan: planned().Plan, Blockers: []string{"tag web@1.0.0 is missing"}})
	if blocked.CanRelease() || press(blocked, "enter").Open {
		t.Fatal("a blocked release offers nothing")
	}
	if out := render(blocked); !strings.Contains(out, "The release is blocked") {
		t.Fatal(out)
	}
	if FromResult(Result{Err: errors.New("no git")}).CanRelease() {
		t.Fatal("a failed plan cannot be released")
	}
	if FromResult(Result{}).CanRelease() {
		t.Fatal("an empty plan has nothing to release")
	}
}

func TestAPlanOnAReleaseChannelSaysItReleasesPrereleases(t *testing.T) {
	m := planned()
	if strings.Contains(render(m), "Release channel") {
		t.Fatal("a stable plan has no channel notice")
	}
	m.Plan.Channel = "next"
	out := render(m)
	for _, want := range []string{"Release channel `next`: prerelease versions", "the stable version stays the latest", "web", "2.0.0"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	if strings.Index(out, "Release channel") > strings.Index(out, "web") {
		t.Fatalf("the notice comes before the versions:\n%s", out)
	}
}

func TestEachRowShowsWhetherItsCommitsAreFoldedOrShown(t *testing.T) {
	// The ASCII icons spell the marks as > (closed) and v (open).
	m := planned()
	out := render(m)
	if strings.Count(out, "> web") != 1 || strings.Count(out, "> kit") != 1 || strings.Contains(out, "v web") {
		t.Fatalf("a closed row shows the closed mark:\n%s", out)
	}
	out = render(press(m, "enter"))
	if !strings.Contains(out, "v web") || !strings.Contains(out, "> kit") {
		t.Fatalf("only the row that is open shows the open mark:\n%s", out)
	}
	lines, _ := planned().View(theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode}), 120)
	if joined := strings.Join(lines, "\n"); !strings.Contains(joined, "▸ web") {
		t.Fatalf("the Unicode theme draws the triangle:\n%s", joined)
	}
}

func TestRightShowsTheCommitsAndLeftHidesThem(t *testing.T) {
	right, left := tea.KeyMsg{Type: tea.KeyRight}, tea.KeyMsg{Type: tea.KeyLeft}
	m := planned().Update(right)
	if !m.Open || !strings.Contains(render(m), "a1b2c3d") {
		t.Fatalf("Right shows the commits:\n%s", render(m))
	}
	if again := m.Update(right); !again.Open {
		t.Fatal("Right on an open row keeps it open")
	}
	m = m.Update(left)
	if m.Open || strings.Contains(render(m), "a1b2c3d") {
		t.Fatalf("Left hides the commits:\n%s", render(m))
	}
	if m = m.Update(left); m.Open || m.Cursor != 0 {
		t.Fatalf("Left on a closed row does nothing: %+v", m)
	}
}

func TestUpAndDownStillMoveWhileTheArrowsOpenAndCloseRows(t *testing.T) {
	m := planned().Update(tea.KeyMsg{Type: tea.KeyRight})
	m = press(m, "down")
	if m.Cursor != 1 || m.Open {
		t.Fatalf("Down moves on and folds the row: %+v", m)
	}
	m = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if !m.Open || m.Cursor != 1 {
		t.Fatalf("Right opens the row the cursor is on: %+v", m)
	}
}
