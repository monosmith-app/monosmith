package guide

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// pathForm is a one-question form that asks for a path, over a folder that
// holds packages/web/tsconfig.json, README.md and dist/.
func pathForm(t *testing.T) Model {
	t.Helper()
	root := t.TempDir()
	for _, f := range []string{"README.md", "packages/web/tsconfig.json", "dist/a.js"} {
		p := filepath.Join(root, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	a := Action{Title: "Why", Steps: []Step{{Key: "path", Prompt: "Which?", Path: true, PathSettings: true}},
		Argv: func(a map[string][]string) []string { return []string{"why", first(a, "path")} }}
	m := NewModel([]Action{a}).WithPathRoot(root)
	return keys(m, "enter").(Model)
}

func TestAPathStepShowsTheFolderAndKeepsTypedPathsWorking(t *testing.T) {
	m := pathForm(t)
	view := m.View()
	for _, want := range []string{"packages/", "README.md", "dist/", "git-ignored", "tab completes"} {
		if !strings.Contains(view, want) {
			t.Errorf("the list lacks %q:\n%s", want, view)
		}
	}
	m = keys(m, "x/y.txt", "enter").(Model)
	if got := strings.Join(m.Argv, " "); got != "why x/y.txt" {
		t.Fatalf("a typed path that is not in the list still works: %q", got)
	}
}

func TestAPathStepTabCompletesAndEnterOpensThenConfirms(t *testing.T) {
	m := pathForm(t)
	m = keys(m, "pack", "tab").(Model)
	if v := m.input.Value(); v != "packages/" {
		t.Fatalf("tab completes the folder: %q", v)
	}
	m = keys(m, "down", "enter").(Model) // the .. row, then web: down once = web? rows are .., web
	if v := m.input.Value(); v != "packages/web/" {
		t.Fatalf("enter on a folder opens it: %q (done %v)", v, m.Done())
	}
	m = keys(m, "down", "enter").(Model)
	if got := strings.Join(m.Argv, " "); got != "why packages/web/tsconfig.json" {
		t.Fatalf("enter on a file confirms it: %q", got)
	}
}

func TestAPathStepSettingsAreOffered(t *testing.T) {
	m := keys(pathForm(t), "devcont").(Model)
	if !strings.Contains(m.View(), "ci.devcontainer") {
		t.Fatalf("the setting is listed:\n%s", m.View())
	}
	m = keys(m, "down", "enter").(Model)
	if got := strings.Join(m.Argv, " "); got != "why ci.devcontainer" {
		t.Fatalf("enter on a setting confirms it: %q", got)
	}
}

func TestAPathStepRowsClickSelectAndFit(t *testing.T) {
	m := pathForm(t)
	lines, optionAt := m.Body(theme.Default(), 80)
	if len(optionAt) == 0 {
		t.Fatal("the rows can be clicked")
	}
	for i, l := range lines {
		if w := len([]rune(stripAnsi(l))); w > 80 {
			t.Errorf("line %d is %d columns", i, w)
		}
	}
	m = m.SelectRow(0)
	if m.cursor != 0 {
		t.Fatalf("cursor %d", m.cursor)
	}
	m, _ = m.Handle(tea.KeyMsg{Type: tea.KeyEnter})
	if v := m.input.Value(); !strings.HasSuffix(v, "/") {
		t.Fatalf("a click then enter opens the folder: %q", v)
	}
}

func TestAskingForPath(t *testing.T) {
	for name, want := range map[string]bool{"file-or-url": true, "sample-folder": true, "cache-dir": true, "recipe-file": true, "extends": true, "name": false, "lang": false, "term": false} {
		if got := asksForPath(name); got != want {
			t.Errorf("asksForPath(%q) = %v", name, got)
		}
	}
}

// stripAnsi removes colour codes so a width can be counted.
func stripAnsi(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		switch {
		case r == 0x1b:
			in = true
		case in && r == 'm':
			in = false
		case !in:
			b.WriteRune(r)
		}
	}
	return b.String()
}
