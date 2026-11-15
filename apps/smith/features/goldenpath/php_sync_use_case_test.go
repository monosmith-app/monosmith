//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPhpSync gives two Composer packages of one workspace different
// constraints on one package: `smith sync --check` fails, `smith sync`
// aligns the composer.json files, and the check then passes. No lockfile
// exists yet, so Composer itself never runs.
func TestPhpSync(t *testing.T) {
	for _, tool := range []string{"php", "composer"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is not installed")
		}
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-phpsync", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-phpsync")
	w.smith(true, "add", "package", "core", "--lang", "php")
	w.smith(true, "add", "app", "cli", "--lang", "php")
	for manifest, constraint := range map[string]string{"packages/core/composer.json": "^3.4.0", "apps/cli/composer.json": "^3.5.0"} {
		w.write(manifest, strings.Replace(w.read(manifest), `"php": ">=8.2"`, `"php": ">=8.2",`+"\n        "+`"monolog/monolog": "`+constraint+`"`, 1))
	}

	out := w.smith(false, "sync", "--check")
	if !strings.Contains(out, "monolog/monolog (php)") || !strings.Contains(out, "^3.5.0") {
		t.Fatalf("sync --check names the package and the constraint it would use:\n%s", out)
	}
	w.smith(true, "sync", "--no-install")
	if got := w.read("packages/core/composer.json"); !strings.Contains(got, `"monolog/monolog": "^3.5.0"`) {
		t.Fatalf("core requires the aligned constraint:\n%s", got)
	}
	if out := w.smith(true, "sync", "--check"); !strings.Contains(out, "Every project agrees") {
		t.Fatalf("after sync the check passes:\n%s", out)
	}
}
