package npmbinaries

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestTheLauncherLogicPassesItsNodeTests runs the node tests of the launcher's
// choices (fake npm, no clock, no terminal). It is skipped where node is not installed.
func TestTheLauncherLogicPassesItsNodeTests(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed")
	}
	out, err := exec.Command(node, "--test", "shim_template.test.js").CombinedOutput()
	if err != nil {
		t.Fatalf("node --test failed:\n%s", strings.TrimSpace(string(out)))
	}
	// The shim that ships is the template with its names filled in: it must parse.
	file := filepath.Join(t.TempDir(), "acme.js")
	if err := os.WriteFile(file, []byte(Shim("acme", map[string]string{"linux-x64": "@acme/cli-linux-x64"})), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(node, "--check", file).CombinedOutput(); err != nil {
		t.Fatalf("the shim does not parse:\n%s", out)
	}
}
