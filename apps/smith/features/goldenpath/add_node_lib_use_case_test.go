//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAddNodeLibInConvertedLayout: in a workspace whose root package.json
// workspaces are packages/* and apps/* (the layout a converted Nx workspace
// has), smith add lib puts the Node project where npm links it, and smith
// doctor finds nothing wrong with it.
func TestAddNodeLibInConvertedLayout(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "nxlayout", "--ci", "github", "--default-stack", "node=node-vitest")
	w.dir = filepath.Join(base, "nxlayout")
	w.write("package.json", `{
  "name": "nxlayout-workspace",
  "private": true,
  "workspaces": ["packages/*", "apps/*"]
}
`)
	w.smith(true, "add", "package", "bar", "--lang", "node")
	out := w.smith(true, "add", "lib", "foo", "--lang", "node")
	if !strings.Contains(out, "Done: lib foo is ready") {
		t.Fatalf("add did not finish:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(w.dir, "packages", "foo", "package.json")); err != nil {
		t.Fatalf("the lib is in the folder the workspaces glob covers: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(w.dir, "libs", "foo")); err == nil {
		t.Fatal("libs/foo must not exist: no workspaces glob covers it")
	}
	if doctor := w.smith(true, "doctor"); strings.Contains(doctor, "is not matched by the root package.json") {
		t.Fatalf("doctor rejects the new lib:\n%s", doctor)
	}
}
