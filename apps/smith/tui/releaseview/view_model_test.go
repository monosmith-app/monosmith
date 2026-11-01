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
