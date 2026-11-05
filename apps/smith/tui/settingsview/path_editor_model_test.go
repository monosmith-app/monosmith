package settingsview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOnlyPathSettingsGetTheFinder(t *testing.T) {
	for path, want := range map[string]bool{"cache.dir": true, "workspace.recipe": true, "cache.remote": false, "workspace.module_prefix": false, "ci.devcontainer": false} {
		if got := takesPath(Setting{Path: path, Type: "string"}); got != want {
			t.Errorf("takesPath(%s) = %v", path, got)
		}
	}
	if takesPath(Setting{Path: "cache.dir", Type: "array"}) {
		t.Error("a list of values is typed, not browsed")
	}
	if newTextEditor(Setting{Path: "ci.provider", Type: "string"}).finder != nil {
		t.Error("a plain text setting has no finder")
	}
}

func TestTheFinderEditsTheValueOfAPathSetting(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "apps"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	m := Model{ed: newTextEditor(Setting{Path: "cache.dir", Type: "string"})}
	for _, k := range []tea.KeyMsg{{Type: tea.KeyRunes, Runes: []rune("apps")}} {
		m, _ = m.editorKey(k)
	}
	if m.ed.finder == nil || !strings.Contains(m.ed.input.Value(), "apps") {
		t.Fatalf("typing still edits the field: %q", m.ed.input.Value())
	}
	m, _ = m.editorKey(tea.KeyMsg{Type: tea.KeyTab})
	if v := m.ed.input.Value(); !strings.HasSuffix(v, "/") {
		t.Fatalf("tab completes the folder: %q", v)
	}
}
