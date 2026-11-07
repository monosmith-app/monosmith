//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestBicepWorkspace builds Bicep with the built-in bicep plugin: a
// deployment using a module from a lib, lint with the linter's errors, the
// ARM templates in dist/, the cache, doctor and CI.
func TestBicepWorkspace(t *testing.T) {
	if _, err := exec.LookPath("az"); err != nil {
		t.Skip("the Azure CLI is not installed")
	}
	if err := exec.Command("az", "bicep", "version").Run(); err != nil {
		t.Skip("the Bicep CLI is not installed: " + err.Error())
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-bicep", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-bicep")
	w.smith(true, "add", "app", "infra", "--lang", "bicep")
	w.smith(true, "add", "lib", "storage", "--lang", "bicep")

	// infra deploys its storage through the lib: a module is a graph edge.
	w.write("apps/infra/main.bicep", strings.Replace(w.read("apps/infra/main.bicep"), "output storageId string = storage.id", `module shared '../../libs/storage/main.bicep' = {
  name: 'shared'
  params: {
    location: location
    storageName: '${storageName}s'
  }
}

output storageId string = storage.id
output sharedId string = shared.outputs.storageId`, 1))
	if !hasEdge(w.smith(true, "graph", "--format", "json"), "infra", "storage") {
		t.Fatal("a module from a lib is a graph edge")
	}
	out := w.smith(true, "run", "lint", "build")
	for _, want := range []string{"infra:lint ✓", "storage:lint ✓", "infra:params ✓", "infra:build ✓", "storage:build ✓"} {
		if !strings.Contains(out, want) {
			t.Fatalf("%s did not pass:\n%s", want, out)
		}
	}
	for _, file := range []string{"apps/infra/template/main.json", "apps/infra/parameters/main.parameters.json", "libs/storage/template/main.json"} {
		if _, err := os.Stat(filepath.Join(w.dir, "dist", filepath.FromSlash(file))); err != nil {
			t.Errorf("dist/%s was not built", file)
		}
	}
	if out := w.smith(true, "run", "lint", "build"); !strings.Contains(out, "(5 from cache") {
		t.Fatalf("a re-run replays from the cache:\n%s", out)
	}
	// The linter's rules are errors: an unused parameter fails lint.
	w.write("libs/storage/main.bicep", w.read("libs/storage/main.bicep")+"\nparam unused string\n")
	if out := w.smith(false, "run", "lint", "-p", "storage"); !strings.Contains(out, "no-unused-params") {
		t.Fatalf("lint fails on an unused parameter:\n%s", out)
	}
	w.write("libs/storage/main.bicep", strings.Replace(w.read("libs/storage/main.bicep"), "\nparam unused string\n", "", 1))
	if out := w.smith(true, "doctor"); !strings.Contains(out, "✓ the Bicep CLI is installed") {
		t.Fatalf("doctor:\n%s", out)
	}
	if ci := w.read(".github/workflows/ci.yml"); !strings.Contains(ci, "az bicep install") {
		t.Fatalf("ci.yml installs Bicep:\n%s", ci)
	}
}
