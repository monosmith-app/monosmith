package pathfinder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func workspaceDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, f := range []string{"README.md", "smith.toml", ".env", "packages/web/src/app.ts", "packages/api/main.go", "dist/out.js", "node_modules/x/index.js", "docs/guide.md"} {
		p := filepath.Join(root, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("# c\n*.log\n/secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func names(m Model) []string {
	var out []string
	for _, r := range m.rows {
		out = append(out, r.name)
	}
	return out
}

func key(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestSyncListsFoldersFirstAndHidesDotFiles(t *testing.T) {
	m := New(workspaceDir(t), nil).Sync("")
	got := strings.Join(names(m), ",")
	if want := "dist,docs,node_modules,packages,README.md,smith.toml"; got != want {
		t.Errorf("rows = %s; want %s", got, want)
	}
	if got := names(New(m.root, nil).Sync(".")); !strings.Contains(strings.Join(got, ","), ".env") {
		t.Errorf("a dot prefix must list hidden files: %v", got)
	}
}

func TestSyncFiltersAndEntersFolders(t *testing.T) {
	m := New(workspaceDir(t), nil).Sync("packages/")
	if got := strings.Join(names(m), ","); got != "..,api,web" {
		t.Errorf("inside packages/ = %s", got)
	}
	if got := strings.Join(names(m.Sync("packages/we")), ","); got != "web" {
		t.Errorf("filter = %s", got)
	}
	if got := strings.Join(names(m.Sync("ack")), ","); got != "packages" {
		t.Errorf("a name that holds the text matches: %s", got)
	}
}

func TestIgnoredAreShownAndFlagged(t *testing.T) {
	m := New(workspaceDir(t), nil).Sync("")
	flagged := map[string]bool{}
	for _, r := range m.rows {
		flagged[r.name] = r.ignored
	}
	if !flagged["dist"] || !flagged["node_modules"] || flagged["docs"] {
		t.Errorf("ignored flags = %v", flagged)
	}
	if !New(m.root, nil).Sync("dist/").rows[1].ignored {
		t.Error("what is inside a git-ignored folder is dimmed too")
	}
}

func TestTabCompletesAndEnterOpensFolder(t *testing.T) {
	m := New(workspaceDir(t), nil).Sync("pack")
	m, out := m.Handle(key("tab"))
	if !out.Set || out.Value != "packages/" {
		t.Fatalf("tab = %+v", out)
	}
	if got := strings.Join(names(m), ","); got != "..,api,web" {
		t.Errorf("after tab the list shows the folder: %s", got)
	}
	m, _ = m.Handle(key("down")) // api
	m, _ = m.Handle(key("down")) // web
	_, out = m.Handle(key("enter"))
	if !out.Set || out.Value != "packages/web/" || out.Confirm {
		t.Errorf("enter on a folder opens it: %+v", out)
	}
}

func TestEnterOnFileConfirmsAndPlainEnterIsLeftToHost(t *testing.T) {
	m := New(workspaceDir(t), nil).Sync("sm")
	if _, out := m.Handle(key("enter")); out.Handled {
		t.Error("enter before moving must be left to the host, which confirms the typed text")
	}
	m, _ = m.Handle(key("down"))
	m = m.Select(0)
	_, out := m.Handle(key("enter"))
	if !out.Confirm || out.Value != "smith.toml" {
		t.Errorf("enter on a file = %+v", out)
	}
}

func TestUpRowGoesUp(t *testing.T) {
	m := New(workspaceDir(t), nil).Sync("packages/web/")
	_, out := m.Select(0).Handle(key("enter"))
	if out.Value != "packages/" || out.Confirm {
		t.Errorf("up row = %+v", out)
	}
}

func TestSettingsAreOfferedAtTheTopLevel(t *testing.T) {
	m := New(workspaceDir(t), []Setting{{Path: "ci.devcontainer", Hint: "a dev container"}, {Path: "cache.dir"}}).Sync("devc")
	if got := strings.Join(names(m), ","); got != "ci.devcontainer" {
		t.Errorf("rows = %s", got)
	}
	if _, out := m.Select(0).Handle(key("enter")); !out.Confirm || out.Value != "ci.devcontainer" {
		t.Errorf("a setting confirms: %+v", out)
	}
	if got := names(New(m.root, m.settings).Sync("")); len(got) > 0 && strings.Contains(strings.Join(got, ","), "ci.") {
		t.Error("settings must not flood an empty prompt")
	}
}

func TestLargeFolderIsCapped(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 300; i++ {
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("f%03d.txt", i)), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	m := New(root, nil).Sync("")
	lines := m.Rows(theme.Default(), 80, 0)
	if len(lines) != 1+DefaultRows+1 {
		t.Fatalf("lines = %d", len(lines))
	}
	if last := lines[len(lines)-1].Text; !strings.Contains(last, "292 more") {
		t.Errorf("last line = %q", last)
	}
	for _, l := range lines {
		if w := len([]rune(l.Text)); w < 1 {
			t.Error("empty line")
		}
	}
}

func TestUnreadableOrEmptyFolderStillAllowsTypedText(t *testing.T) {
	m := New(workspaceDir(t), nil).Sync("nope/")
	if _, out := m.Handle(key("enter")); out.Handled {
		t.Error("enter must go to the host")
	}
	if l := m.Rows(theme.Default(), 80, 0); len(l) != 2 || !strings.Contains(l[1].Text, "cannot be read") {
		t.Errorf("rows = %v", l)
	}
}
