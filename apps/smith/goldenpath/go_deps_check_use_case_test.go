//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGoDepsCheck runs the deps-check task of Go modules in a go.work: a
// fresh module passes, a requirement nothing imports and an import no
// requirement provides fail with the file and the module named, a module that
// imports another go.work member without requiring it passes (go.work
// provides it), and smith doctor reports the same.
func TestGoDepsCheck(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-godepscheck", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-godepscheck")
	w.smith(true, "add", "lib", "left", "--lang", "go")
	w.smith(true, "add", "lib", "right", "--lang", "go")

	if out := w.smith(true, "run", "deps-check", "-p", "left"); !strings.Contains(out, "left") {
		t.Fatalf("a fresh module passes deps-check:\n%s", out)
	}

	original := w.read("libs/left/go.mod")
	w.write("libs/left/go.mod", original+"\nrequire example.invalid/gone v1.0.0\n")
	out := w.smith(false, "run", "deps-check", "-p", "left")
	// Tidy's own diff names it, or the read-only comparison does.
	if !strings.Contains(out, "-require example.invalid/gone v1.0.0") && !strings.Contains(out, "go.mod requires example.invalid/gone, but nothing imports it") {
		t.Fatalf("deps-check names the unused requirement:\n%s", out)
	}
	if out := w.smith(false, "doctor"); !strings.Contains(out, "libs/left/go.mod matches what the module imports") || !strings.Contains(out, "example.invalid/gone") {
		t.Fatalf("doctor reports the unused requirement:\n%s", out)
	}
	w.write("libs/left/go.mod", original)

	w.write("libs/left/extra.go", "package left\n\nimport _ \"example.invalid/absent/pkg\"\n")
	out = w.smith(false, "run", "deps-check", "-p", "left")
	if !strings.Contains(out, "extra.go imports example.invalid/absent/pkg, which no requirement in go.mod provides") {
		t.Fatalf("deps-check names the missing requirement:\n%s", out)
	}
	var sibling string
	for _, line := range strings.Split(w.read("libs/right/go.mod"), "\n") {
		if path, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			sibling = path
		}
	}
	if sibling == "" {
		t.Fatal("the new module has a module line")
	}
	w.write("libs/left/extra.go", "package left\n\nimport _ \""+sibling+"/greeting\"\n")
	if out := w.smith(true, "run", "deps-check", "-p", "left"); !strings.Contains(out, "left") {
		t.Fatalf("a module may import another module go.work uses without requiring it:\n%s", out)
	}
}
