package eject

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestEjectCopiesTheDefaultToTheFileTheToolReads(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".smith/defaults/ruff.toml", "line-length = 100\n")
	target, err := Eject(root, tooldefaults.Ruff, false)
	if err != nil || target != "ruff.toml" {
		t.Fatalf("%q %v", target, err)
	}
	if got := read(t, root, "ruff.toml"); got != "line-length = 100\n" {
		t.Fatalf("copied %q", got)
	}
}

func TestEjectKeepsAFileThatIsThereUnlessForced(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".smith/defaults/mypy.ini", "[mypy]\n")
	write(t, root, "mypy.ini", "[mypy]\nstrict = True\n")
	_, err := Eject(root, tooldefaults.Mypy, false)
	if err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("expected a refusal that names --force: %v", err)
	}
	if got := read(t, root, "mypy.ini"); !strings.Contains(got, "strict") {
		t.Fatalf("the file was overwritten: %q", got)
	}
	if _, err := Eject(root, tooldefaults.Mypy, true); err != nil {
		t.Fatal(err)
	}
	if got := read(t, root, "mypy.ini"); got != "[mypy]\n" {
		t.Fatalf("forced copy: %q", got)
	}
}

func TestEjectingTwiceIsNotAnError(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".smith/defaults/pytest.ini", "[pytest]\n")
	for i := 0; i < 2; i++ {
		if _, err := Eject(root, tooldefaults.Pytest, false); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}
}

func TestEjectWithoutADefaultSaysSo(t *testing.T) {
	_, err := Eject(t.TempDir(), tooldefaults.GolangciLint, false)
	if err == nil || !strings.Contains(err.Error(), "no default for golangci-lint") {
		t.Fatalf("got %v", err)
	}
}

func TestAvailableListsTheDefaultsThatApplyAndWhetherTheTargetIsThere(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".smith/defaults/ruff.toml", "x")
	write(t, root, ".smith/defaults/golangci.yml", "x")
	write(t, root, ".golangci.yml", "y")
	got := Available(root)
	if len(got) != 2 || got[0].Tool.Name != "golangci-lint" || !got[0].Exists || got[1].Tool.Name != "ruff" || got[1].Exists {
		t.Fatalf("%+v", got)
	}
}

func TestResetDefaultsForgetsWhatWasWrittenBefore(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".smith/defaults/ruff.toml", "x")
	write(t, root, ".smith/cache/keep", "x")
	if err := ResetDefaults(root); err != nil {
		t.Fatal(err)
	}
	if got := Available(root); len(got) != 0 {
		t.Fatalf("%+v", got)
	}
	if _, err := os.Stat(filepath.Join(root, ".smith", "cache", "keep")); err != nil {
		t.Fatal("only the defaults are reset")
	}
}

func TestEveryToolHasAUniqueNameAndFile(t *testing.T) {
	names, files := map[string]bool{}, map[string]bool{}
	for _, tool := range tooldefaults.Tools {
		if names[tool.Name] || files[tool.File] || tool.Target == "" {
			t.Errorf("tool %+v", tool)
		}
		names[tool.Name], files[tool.File] = true, true
	}
}
