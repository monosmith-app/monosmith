//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRustSync gives two crates of one Cargo workspace different versions of
// one dependency: `smith sync --check` fails, `smith sync` aligns them, and
// the check then passes. The rewrite keeps the manifests' features.
func TestRustSync(t *testing.T) {
	if _, err := exec.LookPath("cargo"); err != nil {
		t.Skip("cargo is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-syncy", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-syncy")
	w.smith(true, "add", "app", "cli", "--lang", "rust")
	w.smith(true, "add", "package", "core", "--lang", "rust")
	w.write("apps/cli/Cargo.toml", strings.Replace(w.read("apps/cli/Cargo.toml"), "[dependencies]\n", "[dependencies]\nserde = \"1.0.150\"\n", 1))
	w.write("packages/core/Cargo.toml", strings.Replace(w.read("packages/core/Cargo.toml"), "[dependencies]\n", "[dependencies]\nserde = { version = \"1.0.200\", features = [\"derive\"] }\n", 1))

	out := w.smith(false, "sync", "--check")
	if !strings.Contains(out, "serde (rust)") || !strings.Contains(out, "1.0.200") {
		t.Fatalf("sync --check names the dependency and the version it would use:\n%s", out)
	}
	w.smith(true, "sync", "--no-install")
	if got := w.read("apps/cli/Cargo.toml"); !strings.Contains(got, `serde = "1.0.200"`) {
		t.Fatalf("cli declares the aligned version:\n%s", got)
	}
	if got := w.read("packages/core/Cargo.toml"); !strings.Contains(got, `serde = { version = "1.0.200", features = ["derive"] }`) {
		t.Fatalf("core keeps its features:\n%s", got)
	}
	if out := w.smith(true, "sync", "--check"); !strings.Contains(out, "Every project agrees") {
		t.Fatalf("after sync the check passes:\n%s", out)
	}
}
