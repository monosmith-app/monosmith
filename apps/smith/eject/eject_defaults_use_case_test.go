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
	targets, err := Eject(root, tooldefaults.Ruff, false, nil)
	if err != nil || len(targets) != 1 || targets[0] != "ruff.toml" {
		t.Fatalf("%q %v", targets, err)
	}
	if got := read(t, root, "ruff.toml"); got != "line-length = 100\n" {
		t.Fatalf("copied %q", got)
	}
}

func TestEjectKeepsAFileThatIsThereUnlessForced(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".smith/defaults/mypy.ini", "[mypy]\n")
	write(t, root, "mypy.ini", "[mypy]\nstrict = True\n")
	_, err := Eject(root, tooldefaults.Mypy, false, nil)
	if err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("expected a refusal that names --force: %v", err)
	}
	if got := read(t, root, "mypy.ini"); !strings.Contains(got, "strict") {
		t.Fatalf("the file was overwritten: %q", got)
	}
	if _, err := Eject(root, tooldefaults.Mypy, true, nil); err != nil {
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
		if _, err := Eject(root, tooldefaults.Pytest, false, nil); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}
}

func TestEjectWithoutADefaultSaysSo(t *testing.T) {
	_, err := Eject(t.TempDir(), tooldefaults.GolangciLint, false, nil)
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

func TestAPerProjectToolIsCopiedIntoEachProjectThatUsesTheDefault(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".smith/defaults/phpstan.neon", "parameters:\n    level: 8\n")
	write(t, root, "packages/a/src/x.php", "<?php\n")
	write(t, root, "packages/b/src/x.php", "<?php\n")
	targets, err := Eject(root, tooldefaults.PHPStan, false, []string{"packages/a", "packages/b"})
	if err != nil || strings.Join(targets, " ") != "packages/a/phpstan.neon packages/b/phpstan.neon" {
		t.Fatalf("%v %v", targets, err)
	}
	if got := read(t, root, "packages/b/phpstan.neon"); !strings.Contains(got, "level: 8") {
		t.Fatalf("copied %q", got)
	}
	if _, err := Eject(root, tooldefaults.PHPStan, false, nil); err == nil {
		t.Fatal("no project uses the default: nothing to copy")
	}
}

func TestARefusalWritesNothing(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".smith/defaults/phpstan.neon", "default\n")
	write(t, root, "packages/b/phpstan.neon", "mine\n")
	if _, err := Eject(root, tooldefaults.PHPStan, false, []string{"packages/a", "packages/b"}); err == nil {
		t.Fatal("expected a refusal")
	}
	if _, err := os.Stat(filepath.Join(root, "packages", "a", "phpstan.neon")); err == nil {
		t.Fatal("packages/a was written although packages/b refused")
	}
}

func TestReplacementsShowWhatForceWouldLose(t *testing.T) {
	root := t.TempDir()
	tool := tooldefaults.Tool{Name: "demo", File: "demo.cfg", Target: "demo.cfg"}
	def := defaultPath(root, tool)
	if err := os.MkdirAll(filepath.Dir(def), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(def, []byte("a = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Replacements(root, tool, nil); len(got) != 0 {
		t.Fatalf("no file there, nothing is replaced: %v", got)
	}
	if err := os.WriteFile(filepath.Join(root, "demo.cfg"), []byte("a = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Replacements(root, tool, nil); len(got) != 0 {
		t.Fatalf("the same content loses nothing: %v", got)
	}
	if err := os.WriteFile(filepath.Join(root, "demo.cfg"), []byte("a = 2\nmine = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Replacements(root, tool, nil)
	if len(got) != 1 || got[0].Path != "demo.cfg" || !strings.Contains(got[0].Diff, "-mine = true") {
		t.Fatalf("the diff names the line that is lost: %+v", got)
	}
}
