package doctorview

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func sample() Model {
	safe := &protocol.Fix{Description: "restore web's Dockerfile", Safe: true, Files: []protocol.FileEdit{{Path: "apps/web/Dockerfile", Content: "FROM x\n"}}}
	risky := &protocol.Fix{Description: "rewrite api's CI", Files: []protocol.FileEdit{{Path: "ci.yml", Content: "x\n"}}}
	report := doctor.Report{Passed: 1, Failed: 3, Findings: []doctor.Finding{
		{Finding: protocol.Finding{Check: "git is installed", OK: true}, Source: "smith"},
		{Finding: protocol.Finding{Check: "Docker registries of web agree", Detail: "two registries", Remedy: "pick one", Fix: safe}, Source: "docker"},
		{Finding: protocol.Finding{Check: "CI of api matches", Detail: "drift", Remedy: "run smith ci generate", Fix: risky}, Source: "smith"},
		{Finding: protocol.Finding{Check: "Go is installed", Detail: "go is not on PATH", Remedy: "install Go"}, Source: "go"},
	}}
	m := FromResult(Result{Report: report, Projects: []ProjectRef{{Name: "web", Root: "apps/web"}, {Name: "api", Root: "apps/api"}}})
	m.Grouped = true // the workspace's checks first, then web's, then api's
	return m
}

func key(s string) tea.KeyMsg {
	if s == "enter" {
		return tea.KeyMsg{Type: tea.KeyEnter}
	}
	if s == "down" {
		return tea.KeyMsg{Type: tea.KeyDown}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

var plainTheme = theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})

// plain is the page alone in the pane: a check expands in place.
func plain(m Model) string {
	lines, _ := m.View(plainTheme, 120, true)
	return strings.Join(lines, "\n")
}

// plainBeside is the page with the detail beside it, which tells a check in full.
func plainBeside(m Model) string {
	lines, _ := m.View(plainTheme, 120, false)
	return strings.Join(lines, "\n")
}

func TestTheViewGroupsFindingsByProjectWithTheWorkspaceFirst(t *testing.T) {
	out := plain(sample())
	w, a, b := strings.Index(out, "workspace"), strings.Index(out, "web"), strings.Index(out, "api")
	if w < 0 || a < w || b < a {
		t.Fatalf("expected workspace, then web, then api:\n%s", out)
	}
	if !strings.Contains(out, "2 failed") && !strings.Contains(out, "1 failed") {
		t.Fatalf("a group says how many failed:\n%s", out)
	}
}

func TestGTurnsTheGroupingOffAndOn(t *testing.T) {
	m := sample().Update(key("g"))
	if m.Grouped || strings.Contains(plain(m), "workspace") {
		t.Fatalf("flat view still has group headers:\n%s", plain(m))
	}
	if m = m.Update(key("g")); !m.Grouped {
		t.Fatal("g again groups again")
	}
	if FromResult(Result{}).Grouped {
		t.Fatal("the list starts flat")
	}
}

func TestTheSelectedFindingHasAPlannedFixOnlyWhenItCarriesOne(t *testing.T) {
	m := sample() // the cursor is on the first drawn finding: Go is installed, failed, but no fix
	if p, ok := m.SelectedFix(); ok {
		t.Fatalf("a finding without a fix offered %+v", p)
	}
	m = m.Update(key("down")) // web's Docker finding
	p, ok := m.SelectedFix()
	if !ok || p.Description != "restore web's Dockerfile" || p.Ask {
		t.Fatalf("web's fix: %+v ok=%v", p, ok)
	}
	if FixLabel(p) != "Fix: restore web's Dockerfile" {
		t.Fatalf("label %q", FixLabel(p))
	}
	m = m.Update(key("down")) // api's finding: its fix is not safe, so the wizard asks
	if p, ok := m.SelectedFix(); !ok || !p.Ask {
		t.Fatalf("a fix that is not safe is still a fix, and asks: %+v ok=%v", p, ok)
	}
}

func TestFixAllStartsOnTheFixesThatNeedNoDecisionElseOnTheFirstThatAsks(t *testing.T) {
	m := sample()
	p, ok := m.FixAll()
	if !ok || p.Description != "restore web's Dockerfile" {
		t.Fatalf("%+v ok=%v", p, ok)
	}
	m.Report.Findings[1].Fix = nil
	m.Report.Fixes = nil
	p, ok = m.FixAll()
	if !ok || !p.Ask {
		t.Fatalf("only the fix that asks is left: %+v ok=%v", p, ok)
	}
	m.Report.Findings[2].Fix = nil
	if _, ok := m.FixAll(); ok {
		t.Fatal("nothing to fix, yet a fix was offered")
	}
}

func TestAFixFoldedFromManyProjectsIsLabelledByTheirNumber(t *testing.T) {
	p := doctorfix.Planned{Fix: protocol.Fix{Description: "declare imports", AppliesTo: []string{"a", "b", "c"}}}
	if got := FixLabel(p); got != "Fix all 3 projects" {
		t.Fatalf("label %q", got)
	}
}

func TestOpeningAFindingShowsItsFixAndWhetherItIsSafe(t *testing.T) {
	m := sample()
	m = m.Update(key("down"))
	out := plain(m.Update(key("enter")))
	for _, want := range []string{"→ pick one", "fix: restore web's Dockerfile", "f shows what changes, then applies it", "writes apps/web/Dockerfile"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	m = m.Update(key("down")).Update(key("enter"))
	if out := plain(m); !strings.Contains(out, "f asks, shows what changes") {
		t.Fatalf("a fix that is not safe says so:\n%s", out)
	}
}

func TestPassedChecksAreOneCollapsedLineThatEnterOrPOpens(t *testing.T) {
	m := sample()
	out := plain(m)
	if strings.Contains(out, "git is installed") || !strings.Contains(out, "1 checks passed") {
		t.Fatalf("passes should be collapsed:\n%s", out)
	}
	if strings.Contains(out, "install Go") || strings.Contains(out, "go is not on PATH") {
		t.Fatalf("a check is one line until it is expanded:\n%s", out)
	}
	// Three problems, then the collapsed line: enter on it opens the passes.
	for range 3 {
		m = m.Update(key("down"))
	}
	if !m.OnPassedLine() {
		t.Fatal("the cursor should reach the collapsed line")
	}
	if m = m.Update(key("enter")); !m.ShowPassed || !strings.Contains(plain(m), "git is installed") {
		t.Fatal("enter opens the passed checks")
	}
	if m = m.Update(key("p")); m.ShowPassed {
		t.Fatal("p hides them again")
	}
}

func TestSelectRowFollowsTheFindingEvenWhenGrouped(t *testing.T) {
	m := sample().SelectRow(1) // web's Docker finding, drawn third
	if f, _ := m.Selected(); f.Check != "Docker registries of web agree" {
		t.Fatalf("selected %q", f.Check)
	}
	if !m.SelectRow(1).Open {
		t.Fatal("selecting the selected row again opens it")
	}
}

func TestFailuresAreListedBeforeWarnings(t *testing.T) {
	report := doctor.Report{Failed: 1, Warnings: 1, Findings: []doctor.Finding{
		{Finding: protocol.Finding{Check: "a warned thing", Warning: true, Detail: "d", Remedy: "r"}, Source: "smith"},
		{Finding: protocol.Finding{Check: "a failed thing", Detail: "d", Remedy: "r"}, Source: "smith"},
	}}
	m := FromResult(Result{Report: report})
	out := plain(m)
	if f, w := strings.Index(out, "a failed thing"), strings.Index(out, "a warned thing"); f < 0 || w < f {
		t.Fatalf("the failure comes first:\n%s", out)
	}
	if got, _ := m.Selected(); got.Check != "a failed thing" {
		t.Fatalf("the cursor starts on the failure, got %q", got.Check)
	}
}

func TestTheHeadlineComesFirstThenOneLinePerFailingCheck(t *testing.T) {
	flat := sample()
	flat.Grouped = false
	lines, _ := flat.View(plainTheme, 120, true)
	if !strings.Contains(lines[0], "3 failed") {
		t.Fatalf("the headline is first: %q", lines[0])
	}
	var checks []string
	for _, l := range lines[2:] {
		if strings.TrimSpace(l) != "" && !strings.Contains(l, "checks passed") {
			checks = append(checks, l)
		}
	}
	if len(checks) != 3 {
		t.Fatalf("one line per failing check, got %d:\n%s", len(checks), strings.Join(lines, "\n"))
	}
	if !strings.Contains(checks[0], "Docker registries of web agree") {
		t.Fatalf("a line is the state and the title: %q", checks[0])
	}
}

func TestEnterExpandsOneCheckAndTheProblemAndRemedyShowOnce(t *testing.T) {
	m := sample().Update(key("enter"))
	out := plain(m)
	if strings.Count(out, "go is not on PATH") != 1 || strings.Count(out, "install Go") != 1 {
		t.Fatalf("problem and remedy once:\n%s", out)
	}
	if strings.Contains(out, "two registries") {
		t.Fatalf("only the selected check expands:\n%s", out)
	}
	m = m.Update(key("down"))
	if strings.Contains(plain(m), "go is not on PATH") {
		t.Fatal("moving closes the expanded check")
	}
}

func TestBesideTheDetailTheListDoesNotRepeatWhatTheDetailSays(t *testing.T) {
	out := plainBeside(sample().Update(key("enter")))
	if strings.Contains(out, "go is not on PATH") || strings.Contains(out, "install Go") {
		t.Fatalf("the detail pane tells it, the list does not:\n%s", out)
	}
}

func TestTextTellsTheProblemTheRemedyAndTheFixOnce(t *testing.T) {
	f, _ := sample().Update(key("down")).Selected()
	got := strings.Join(Text(plainTheme, f, 2), "\n")
	for _, want := range []string{"two registries", "pick one", "fix: restore web's Dockerfile", "writes apps/web/Dockerfile"} {
		if strings.Count(got, want) != 1 {
			t.Errorf("%q should appear once:\n%s", want, got)
		}
	}
}
