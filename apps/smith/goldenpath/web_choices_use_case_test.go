//go:build e2e

package goldenpath

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// selections reads testdata/web-selections.txt: one --with selection per line,
// blank lines and # comments ignored. Adding a line adds a case; a choice that is
// only data (ADR 0045) needs no Go code.
func selections(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "web-selections.txt"))
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, line := range strings.Split(string(raw), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			out = append(out, line)
		}
	}
	return out
}

// TestWebChoices builds a browser app from each selection of choices in
// testdata/web-selections.txt (smith add app <name> --with <selection>) and runs
// lint, typecheck, test and build, then the doctor. It also checks that an invalid
// selection gives one message with the fix, and that a ready-made stack, which is
// a saved selection, still works.
func TestWebChoices(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-choices", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-choices")

	if out := w.smith(true, "choices"); !strings.Contains(out, "framework") || !strings.Contains(out, "style") || !strings.Contains(out, "test") {
		t.Fatalf("smith choices lists the axes:\n%s", out)
	}
	for _, bad := range []struct{ with, want string }{
		{"react,vue", "react and vue are both on the framework axis: keep one"},
		{"vue,p5", "p5 requires react, which is not in the selection: add react to --with, or drop p5"},
		{"bootstrap", "no choice in bootstrap writes the project"},
	} {
		if out := w.smith(false, "add", "app", "bad", "--with", bad.with); !strings.Contains(out, bad.want) {
			t.Fatalf("--with %s:\n%s\nwant: %s", bad.with, out, bad.want)
		}
	}
	if _, err := os.Stat(filepath.Join(w.dir, "apps", "bad")); err == nil {
		t.Fatal("a refused selection wrote files")
	}

	names := map[string]string{}
	for _, with := range selections(t) {
		name := strings.ReplaceAll(with, ",", "-")
		names[name] = with
		w.smith(true, "add", "app", name, "--with", with)
	}
	// A ready-made stack is a saved selection and still works.
	w.smith(true, "add", "app", "plain-react", "--stack", "react")
	names["plain-react"] = "react"

	out := w.smith(true, "run", "lint", "typecheck", "test", "build")
	for name := range names {
		for _, task := range []string{"lint", "typecheck", "test", "build"} {
			if want := name + ":" + task + " ✔"; !strings.Contains(out, want) {
				t.Fatalf("%s did not pass:\n%s", want, out)
			}
		}
		if _, err := os.Stat(filepath.Join(w.dir, "dist", "apps", name, "index.html")); err != nil {
			t.Errorf("%s did not build into dist/apps/%s", name, name)
		}
	}
	if doctor := w.smith(true, "doctor"); !strings.Contains(doctor, "0 failed") {
		t.Fatalf("doctor:\n%s", doctor)
	}
}
