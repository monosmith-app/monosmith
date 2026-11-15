//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDartSync gives two Dart packages of one pub workspace different
// constraints on one dependency: `smith sync --check` fails, `smith sync`
// aligns them, and the check then passes.
func TestDartSync(t *testing.T) {
	if _, err := exec.LookPath("dart"); err != nil {
		t.Skip("dart is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-dartsync", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-dartsync")
	w.smith(true, "add", "package", "core", "--lang", "dart")
	w.smith(true, "add", "app", "tool", "--lang", "dart")
	for pubspec, constraint := range map[string]string{"packages/core/pubspec.yaml": "^1.1.0", "apps/tool/pubspec.yaml": "^1.2.0"} {
		w.write(pubspec, strings.Replace(w.read(pubspec), "\ndev_dependencies:", "\ndependencies:\n  http: "+constraint+"\n\ndev_dependencies:", 1))
	}

	out := w.smith(false, "sync", "--check")
	if !strings.Contains(out, "http (dart)") || !strings.Contains(out, "^1.2.0") {
		t.Fatalf("sync --check names the package and the constraint it would use:\n%s", out)
	}
	w.smith(true, "sync", "--no-install")
	if got := w.read("packages/core/pubspec.yaml"); !strings.Contains(got, "  http: ^1.2.0\n") {
		t.Fatalf("core declares the aligned constraint:\n%s", got)
	}
	if out := w.smith(true, "sync", "--check"); !strings.Contains(out, "Every project agrees") {
		t.Fatalf("after sync the check passes:\n%s", out)
	}
}
