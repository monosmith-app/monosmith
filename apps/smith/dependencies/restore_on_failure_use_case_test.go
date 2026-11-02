package dependencies

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAFailedInstallPutsTheManifestAndLockBack(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "apps", "api-py")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(project, "pyproject.toml")
	lock := filepath.Join(root, "uv.lock")
	write(t, manifest, "[project]\ndependencies = []\n")
	write(t, lock, "version = 1\n")

	// The fake installer does what pip does: it writes the manifest, then fails.
	failure := errors.New("No matching distribution found for core-py")
	err := guarded([]string{project, root}, func() error {
		write(t, manifest, "[project]\ndependencies = [\"core-py\"]\n")
		write(t, lock, "version = 2\n")
		return failure
	})

	if !errors.Is(err, failure) {
		t.Fatalf("the failure must stay the cause: %v", err)
	}
	if !strings.Contains(err.Error(), "put pyproject.toml, uv.lock back as it was") {
		t.Errorf("the message must say what came back: %v", err)
	}
	if got := read(t, manifest); got != "[project]\ndependencies = []\n" {
		t.Errorf("manifest %q", got)
	}
	if got := read(t, lock); got != "version = 1\n" {
		t.Errorf("lock %q", got)
	}
}

func TestASuccessfulInstallKeepsItsChanges(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "package.json")
	write(t, manifest, "{}")
	if err := guarded([]string{dir}, func() error { write(t, manifest, `{"dependencies":{}}`); return nil }); err != nil {
		t.Fatal(err)
	}
	if got := read(t, manifest); got != `{"dependencies":{}}` {
		t.Errorf("manifest %q", got)
	}
}

func TestAFailureThatChangedNothingKeepsItsOwnMessage(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), "module x\n")
	failure := errors.New("proxy unreachable")
	if err := guarded([]string{dir}, func() error { return failure }); err != failure {
		t.Fatalf("%v", err)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
