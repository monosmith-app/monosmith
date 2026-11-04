//go:build e2e

package goldenpath

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWebWorkspace scaffolds a browser app with each web stack (the vite
// plugin's, angular, docs-site) and runs every task: lint, typecheck, test
// and the build into dist/, then a cache replay.
func TestWebWorkspace(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-web", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-web")
	all := []string{"lint", "typecheck", "test", "build"}
	stacks := map[string][]string{
		"react": all, "react-p5": all, "vue": all, "svelte": all, "html": all, "angular": all,
		"docs-site": {"lint", "build"},
	}
	for stack := range stacks {
		w.smith(true, "add", "app", stack, "--stack", stack)
	}
	if out := w.smith(false, "add", "lib", "ui", "--stack", "react"); !strings.Contains(out, "cannot scaffold a lib; it scaffolds app") {
		t.Fatalf("a web stack scaffolds apps only:\n%s", out)
	}

	out := w.smith(true, "run", "lint", "typecheck", "test", "build")
	for stack, tasks := range stacks {
		for _, task := range tasks {
			if want := stack + ":" + task + " ✔"; !strings.Contains(out, want) {
				t.Fatalf("%s did not pass:\n%s", want, out)
			}
		}
		// Builds go to dist/, never dist/ in the app.
		if _, err := os.Stat(filepath.Join(w.dir, "dist", "apps", stack, "index.html")); err != nil {
			t.Errorf("%s did not build into dist/apps/%s", stack, stack)
		}
		if _, err := os.Stat(filepath.Join(w.dir, "apps", stack, "dist")); err == nil {
			t.Errorf("apps/%s/dist exists: builds must go to the workspace's dist/", stack)
		}
	}
	if out := w.smith(true, "run", "lint", "typecheck", "test", "build"); !strings.Contains(out, "(26 from cache") {
		t.Fatalf("a re-run replays from the cache:\n%s", out)
	}
	if out := w.smith(true, "stacks"); !strings.Contains(out, "react-p5") || !strings.Contains(out, "docs-site") {
		t.Fatalf("the vite plugin's stacks are listed:\n%s", out)
	}
}
