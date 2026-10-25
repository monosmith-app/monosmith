//go:build e2e

package goldenpath

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConvertTurborepoFullLevel converts the pnpm Turborepo of
// testdata/turborepo-pnpm at the full level, after giving it the leftovers of
// a Rollup build: the plan names them and the flags that apply it, a run
// without a terminal refuses to change anything, and --yes removes the
// config, takes the package out of package.json, and leaves a workspace that
// builds and tests.
func TestConvertTurborepoFullLevel(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: filepath.Join(base, "turbo"), env: identity()}
	copyTree(t, filepath.Join("testdata", "turborepo-pnpm"), w.dir)
	if err := os.WriteFile(filepath.Join(w.dir, "packages", "util", "rollup.config.cjs"), []byte("module.exports = {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := w.read("package.json")
	root = strings.Replace(root, `"devDependencies": {`, `"devDependencies": {
    "rollup": "^4.14.0",`, 1)
	if err := os.WriteFile(filepath.Join(w.dir, "package.json"), []byte(root), 0o644); err != nil {
		t.Fatal(err)
	}
	w.git("init", "-q", "-b", "main")
	w.git("add", "-A")
	w.git("commit", "-qm", "chore: turborepo with a rollup leftover", "--no-verify")

	plan := w.smith(true, "convert", "--level", "full", "--build", "smith", "--dry-run")
	for _, want := range []string{
		"Full conversion: smith becomes the only tool",
		"Remove 1 build package and 1 build config file of the old build",
		"rollup.config.cjs",
		"smith convert --level full --lint keep --test",
		"Nothing changed (dry run)",
	} {
		if !strings.Contains(plan, want) {
			t.Fatalf("plan lacks %q:\n%s", want, plan)
		}
	}

	// Without a terminal and without --yes nothing may change.
	out := w.smith(false, "convert", "--level", "full")
	if !strings.Contains(out, "needs a terminal: nothing was changed") || !strings.Contains(out, "--yes") {
		t.Fatalf("a run that cannot ask must say so and name the flags:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(w.dir, "smith.toml")); err == nil {
		t.Fatal("the refused run changed the workspace")
	}

	w.smith(true, "convert", "--level", "full", "--build", "smith", "--yes")
	if _, err := os.Stat(filepath.Join(w.dir, "packages", "util", "rollup.config.cjs")); err == nil {
		t.Fatal("the Rollup config must be gone")
	}
	if got := w.read("package.json"); strings.Contains(got, "rollup") {
		t.Fatalf("package.json still depends on rollup:\n%s", got)
	}
	if out := w.smith(true, "run", "build", "test"); !strings.Contains(out, "tasks succeeded") {
		t.Fatal(out)
	}
}
