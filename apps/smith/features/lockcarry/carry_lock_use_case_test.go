package lockcarry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func put(t *testing.T, dir, rel, text string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCarryNpmWritesTheSourceVersionsIntoTheWorkspaceLock(t *testing.T) {
	root, src := t.TempDir(), t.TempDir()
	put(t, root, "package-lock.json", workspaceLock)
	put(t, root, "apps/web/package.json", manifest)
	put(t, src, "package-lock.json", `{"lockfileVersion": 3, "packages": {"node_modules/react-icons": {"version": "4.3.1"}}}`)
	r, err := CarryNpm(root, src, "apps/web")
	if err != nil {
		t.Fatal(err)
	}
	if r.Lockfile != "package-lock.json" || r.Added != 1 || len(r.Kept) != 1 {
		t.Errorf("result = %+v", r)
	}
	raw, _ := os.ReadFile(filepath.Join(root, "package-lock.json"))
	if !strings.Contains(string(raw), `"apps/web/node_modules/react-icons"`) {
		t.Error("the workspace lockfile does not hold the project's entry")
	}
}

func TestCarryNpmNamesALockfileItCannotCarry(t *testing.T) {
	root, src := t.TempDir(), t.TempDir()
	put(t, src, "yarn.lock", "")
	r, err := CarryNpm(root, src, "apps/web")
	if err != nil || r.Unsupported != "yarn.lock" || r.Lockfile != "" {
		t.Errorf("result = %+v, %v", r, err)
	}
}

func TestCarryNpmSaysWhenTheWorkspaceHasNoLockfile(t *testing.T) {
	root, src := t.TempDir(), t.TempDir()
	put(t, src, "package-lock.json", `{"lockfileVersion": 3, "packages": {}}`)
	r, err := CarryNpm(root, src, "apps/web")
	if err != nil || !r.NoWorkspaceLock {
		t.Errorf("result = %+v, %v", r, err)
	}
}
