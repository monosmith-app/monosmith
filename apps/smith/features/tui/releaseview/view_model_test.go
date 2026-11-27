package releaseview

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
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

var plain = theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})

func render(m Model) string {
	lines, _ := m.View(plain, 120)
	return strings.Join(lines, "\n")
}

// detailText is the detail of the selected package as plain text.
func detailText(m Model) string {
	var out []string
	for _, s := range m.Detail(plain) {
		out = append(out, s.Heading)
		out = append(out, s.Items...)
	}
	return pane.Strip(strings.Join(out, "\n"))
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

func TestEachProjectIsARowOfThePlanTable(t *testing.T) {
	out := render(planned())
	for _, want := range []string{"Package", "Now", "Next", "Bump", "Why", "web", "1.0.0", "2.0.0", "major", "kit", "0.2.1", "patch", "unchanged: sdk", "2 packages to release · 1 major · 1 patch"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "a1b2c3d") {
		t.Fatalf("the commits are in the detail, not in the list:\n%s", out)
	}
}

func TestTheDetailShowsTheCommitsBehindTheBump(t *testing.T) {
	out := detailText(planned())
	for _, want := range []string{"Why it releases", "3 releasing commit(s)", "Changes", "a1b2c3d feat(ui,theme): dark mode", "Ada", "9f8e7d6 fix!: cookie parsing", "BREAKING: drops the old format", "Tag", "web@2.0.0", "The same as", "$ smith release --dry-run --notes"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "a1b2c3d4e5f6") {
		t.Fatalf("the hash is shortened like git does:\n%s", out)
	}
	if got := planned().DetailTitle(); got != "web 1.0.0 → 2.0.0" {
		t.Fatalf("title %q", got)
	}
}

func TestADependencyBumpSaysNoCommitOfItsOwnReleasesIt(t *testing.T) {
	out := detailText(press(planned(), "down"))
	if !strings.Contains(out, "dependency web releases (no commit of its own releases it)") {
		t.Fatalf("%s", out)
	}
}

func TestMovingWrapsAroundAndAClickOnlySelects(t *testing.T) {
	m := press(planned(), "down")
	if m.Cursor != 1 {
		t.Fatalf("cursor %d", m.Cursor)
	}
	if m = press(m, "down"); m.Cursor != 0 {
		t.Fatalf("down from the last wraps to the first: %d", m.Cursor)
	}
	if m = m.SelectRow(1); m.Cursor != 1 || m.Open {
		t.Fatalf("a click selects: %d %v", m.Cursor, m.Open)
	}
	if m = m.SelectRow(1); m.Open {
		t.Fatal("a second click selects again; the shell opens the detail")
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

func TestOnlyTheSelectedRowHasTheCursorMark(t *testing.T) {
	// The ASCII icons spell the cursor mark as >.
	out := render(planned())
	if strings.Count(out, "> web") != 1 || strings.Contains(out, "> kit") {
		t.Fatalf("only the selected row has the mark:\n%s", out)
	}
	out = render(press(planned(), "down"))
	if strings.Contains(out, "> web") || !strings.Contains(out, "> kit") {
		t.Fatalf("the mark follows the cursor:\n%s", out)
	}
}

func TestNoKeyOpensTheCommitsInTheList(t *testing.T) {
	for _, k := range []tea.KeyMsg{{Type: tea.KeyRight}, {Type: tea.KeyLeft}, {Type: tea.KeyEnter}} {
		m := planned().Update(k)
		if m.Open || m.Cursor != 0 || strings.Contains(render(m), "a1b2c3d") {
			t.Fatalf("%s opened the commits in the list: %+v", k, m)
		}
	}
}

func TestTheWaitShowsASpinner(t *testing.T) {
	m := Model{Tick: 1}
	lines, _ := m.View(theme.Default(), 80)
	if got := pane.Strip(strings.Join(lines, " ")); !strings.Contains(got, "Reading the commits") || strings.TrimSpace(got)[0] == 'R' {
		t.Fatalf("the wait starts with the spinner: %q", got)
	}
}

func TestTheBumpsHaveTheirMarks(t *testing.T) {
	th := theme.Default()
	for level, want := range map[string]string{"patch": th.Icons.Dot + " patch", "minor": th.Icons.Dot + " minor", "major": th.Icons.Warning + " major"} {
		if got := pane.Strip(BumpMark(th, level)); got != want {
			t.Errorf("%s: %q, want %q", level, got, want)
		}
	}
}
