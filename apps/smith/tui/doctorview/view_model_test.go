package doctorview

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
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
	return FromResult(Result{Report: report, Projects: []ProjectRef{{Name: "web", Root: "apps/web"}, {Name: "api", Root: "apps/api"}}})
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

func plain(m Model) string {
	lines, _ := m.View(theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII}), 120)
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
	if !m.Flat || strings.Contains(plain(m), "workspace") {
		t.Fatalf("flat view still has group headers:\n%s", plain(m))
	}
	if m = m.Update(key("g")); m.Flat {
		t.Fatal("g again groups again")
	}
}

func TestFixArgvIsTheSelectedFindingsSafeFixAndNothingElse(t *testing.T) {
	m := sample() // the cursor is on the first drawn finding: Go is installed, failed, but no fix
	if got := m.FixArgv(); got != nil {
		t.Fatalf("a finding without a fix offered %v", got)
	}
	m = m.Update(key("down")) // web's Docker finding
	f, _ := m.Selected()
	if f.Check != "Docker registries of web agree" {
		t.Fatalf("selected %q", f.Check)
	}
	want := []string{"doctor", "--fix", "--only", "Docker registries of web agree"}
	if got := m.FixArgv(); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	m = m.Update(key("down")) // api's finding: its fix is not safe
	if got := m.FixArgv(); got != nil {
		t.Fatalf("a fix that is not safe was offered: %v", got)
	}
}

func TestOpeningAFindingShowsItsFixAndWhetherItIsSafe(t *testing.T) {
	m := sample()
	m = m.Update(key("down"))
	out := plain(m.Update(key("enter")))
	for _, want := range []string{"→ pick one", "fix: restore web's Dockerfile", "f applies it", "writes apps/web/Dockerfile"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	m = m.Update(key("down")).Update(key("enter"))
	if out := plain(m); !strings.Contains(out, "never applied") {
		t.Fatalf("a fix that is not safe says so:\n%s", out)
	}
}

func TestPassedChecksAreOneCollapsedLineThatEnterOrPOpens(t *testing.T) {
	m := sample()
	out := plain(m)
	if strings.Contains(out, "git is installed") || !strings.Contains(out, "1 checks passed") {
		t.Fatalf("passes should be collapsed:\n%s", out)
	}
	if !strings.Contains(out, "→ install Go") || !strings.Contains(out, "go is not on PATH") {
		t.Fatalf("a problem shows its detail and remedy inline:\n%s", out)
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
