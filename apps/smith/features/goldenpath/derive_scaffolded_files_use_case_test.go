//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDeriveRestoresTheFilesASelectionScaffolded is the exit test of issue 840
// (ADR 0049): in a workspace made by smith new and smith add, with derive on,
// the tsconfig.json of every Node project is in smith.lock from the add. A
// deleted one comes back whole and identical from smith derive, the tasks stay
// green, and a hand edit is reported as drift and never overwritten.
func TestDeriveRestoresTheFilesASelectionScaffolded(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "dm", "--ci", "github")
	w.dir = filepath.Join(base, "dm")
	w.write("smith.toml", w.read("smith.toml")+"\n[derive]\nenabled = true\n")

	w.smith(true, "add", "app", "shop", "--with", "react,vitest")
	w.smith(true, "add", "app", "site", "--with", "vue,jest,tailwind")
	w.smith(true, "add", "lib", "core", "--lang", "node")

	files := []string{"apps/shop/tsconfig.json", "apps/site/tsconfig.json", "libs/core/tsconfig.json"}
	lock := w.read("smith.lock")
	want := map[string]string{}
	for _, f := range files {
		if !strings.Contains(lock, "path = '"+f+"'") {
			t.Fatalf("smith add did not record %s in smith.lock:\n%s", f, lock)
		}
		want[f] = w.read(f)
	}
	if out := w.smith(true, "derive", "--check"); !strings.Contains(out, "3 derived files are up to date") {
		t.Fatalf("a fresh workspace is not up to date:\n%s", out)
	}

	for _, f := range files {
		if err := os.Remove(filepath.Join(w.dir, filepath.FromSlash(f))); err != nil {
			t.Fatal(err)
		}
	}
	if out := w.smith(false, "derive", "--check"); !strings.Contains(out, "out of date") {
		t.Fatalf("--check does not fail on deleted files:\n%s", out)
	}
	w.smith(true, "derive")
	for _, f := range files {
		if got := w.read(f); got != want[f] {
			t.Fatalf("%s did not come back as it was:\n%s\nwant:\n%s", f, got, want[f])
		}
	}
	w.smith(true, "run", "build", "test")
	w.smith(true, "doctor")

	edited := strings.Replace(want[files[2]], `"rootDir": "src"`, `"rootDir": "src", "strict": true`, 1)
	w.write(files[2], edited)
	out := w.smith(false, "derive")
	if !strings.Contains(out, "drift") || !strings.Contains(out, files[2]) {
		t.Fatalf("a hand edit is not reported as drift:\n%s", out)
	}
	if w.read(files[2]) != edited {
		t.Fatal("derive overwrote a hand edit")
	}
}
