//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGoSync gives two Go modules of one go.work different versions of one
// dependency: `smith sync --check` fails, `smith sync` aligns them with go
// get (go.sum follows), and the check then passes.
func TestGoSync(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-gosync", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-gosync")
	w.smith(true, "add", "lib", "left", "--lang", "go")
	w.smith(true, "add", "lib", "right", "--lang", "go")
	for module, version := range map[string]string{"libs/left": "v1.5.0", "libs/right": "v1.6.0"} {
		w.write(module+"/go.mod", w.read(module+"/go.mod")+"\nrequire github.com/google/uuid "+version+"\n")
	}

	out := w.smith(false, "sync", "--check")
	if !strings.Contains(out, "github.com/google/uuid (go)") || !strings.Contains(out, "1.6.0") {
		t.Fatalf("sync --check names the module and the version it would use:\n%s", out)
	}
	w.smith(true, "sync")
	for _, module := range []string{"libs/left", "libs/right"} {
		if got := w.read(module + "/go.mod"); !strings.Contains(got, "github.com/google/uuid v1.6.0") {
			t.Fatalf("%s declares the aligned version:\n%s", module, got)
		}
	}
	if out := w.smith(true, "sync", "--check"); !strings.Contains(out, "Every project agrees") {
		t.Fatalf("after sync the check passes:\n%s", out)
	}
}
