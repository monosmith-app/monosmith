//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRubySync gives two gems of one Bundler workspace different constraints
// on one gem: `smith sync --check` fails, `smith sync` aligns the gemspecs,
// and the check then passes.
func TestRubySync(t *testing.T) {
	for _, tool := range []string{"ruby", "bundle"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is not installed")
		}
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-rubysync", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-rubysync")
	w.smith(true, "add", "package", "core", "--lang", "ruby")
	w.smith(true, "add", "app", "cli", "--lang", "ruby")
	for spec, constraint := range map[string]string{
		"packages/core/smith_e2e_rubysync_core.gemspec": "~> 2.2.0",
		"apps/cli/smith_e2e_rubysync_cli.gemspec":       "~> 2.3.0",
	} {
		w.write(spec, strings.Replace(w.read(spec), "\nend", "\n  spec.add_dependency 'rack', '"+constraint+"'\nend", 1))
	}

	out := w.smith(false, "sync", "--check")
	if !strings.Contains(out, "rack (ruby)") || !strings.Contains(out, "~> 2.3.0") {
		t.Fatalf("sync --check names the gem and the constraint it would use:\n%s", out)
	}
	w.smith(true, "sync", "--no-install")
	if got := w.read("packages/core/smith_e2e_rubysync_core.gemspec"); !strings.Contains(got, "spec.add_dependency 'rack', '~> 2.3.0'") {
		t.Fatalf("core declares the aligned constraint:\n%s", got)
	}
	if out := w.smith(true, "sync", "--check"); !strings.Contains(out, "Every project agrees") {
		t.Fatalf("after sync the check passes:\n%s", out)
	}
}
