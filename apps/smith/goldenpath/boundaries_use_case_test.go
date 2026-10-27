//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const adminGreeting = "package greeting\n\n// Hello is what the admin app says.\nfunc Hello() string { return \"hello\" }\n"

const webImportsAdmin = "package main\n\nimport (\n\t\"fmt\"\n\n\tadmin \"example.com/bndgo/apps/admin/greeting\"\n)\n\nfunc main() { fmt.Println(admin.Hello()) }\n"

const webImportsNothing = "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"web\") }\n"

// TestBoundaries: an app that imports another app is reported by doctor and
// smith boundaries with the edge and the rule, fails smith ci verify once the
// rule is an error, and is clear again after an exception or after the import
// is removed.
func TestBoundaries(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "bndgo", "--ci", "github")
	w.dir = filepath.Join(base, "bndgo")
	w.smith(true, "add", "app", "web", "--lang", "go")
	w.smith(true, "add", "app", "admin", "--lang", "go")
	w.write("apps/admin/greeting/greeting.go", adminGreeting)
	w.write("apps/web/main.go", webImportsAdmin)
	config := w.read("smith.toml")

	// Nothing set: a warning that names the edge, and nothing fails.
	out := w.smith(true, "boundaries")
	for _, want := range []string{"web -> admin", "an app is deployed", "0 error(s), 1 warning(s)"} {
		if !strings.Contains(out, want) {
			t.Fatalf("boundaries lacks %q:\n%s", want, out)
		}
	}
	if doctor := w.smith(true, "doctor"); !strings.Contains(doctor, "no project depends on an app") {
		t.Fatalf("doctor does not report the edge:\n%s", doctor)
	}

	// kinds = true: an error everywhere, with the edge, the rule and the hint.
	w.write("smith.toml", config+"\n[boundaries]\nkinds = true\n")
	out = w.smith(false, "boundaries")
	for _, want := range []string{"web -> admin", "rule: an app is deployed", "remove web's import of admin"} {
		if !strings.Contains(out, want) {
			t.Fatalf("boundaries lacks %q:\n%s", want, out)
		}
	}
	if doctor := w.smith(false, "doctor"); !strings.Contains(doctor, "web -> admin") {
		t.Fatalf("doctor does not name the edge:\n%s", doctor)
	}
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: web and admin")
	if verify := w.smith(false, "ci", "verify"); !strings.Contains(verify, "web -> admin") {
		t.Fatalf("ci verify does not name the edge:\n%s", verify)
	}

	// A rule by tag, with the built-in one off.
	w.write("smith.toml", config+`
[boundaries]
kinds = false

[projects.web]
tags = ["ui"]

[projects.admin]
tags = ["server"]

[[boundaries.rules]]
from = "tag:ui"
deny = ["tag:server"]
reason = "the browser must not pull in server code"
`)
	out = w.smith(false, "boundaries")
	if !strings.Contains(out, "web -> admin") || !strings.Contains(out, "the browser must not pull in server code") {
		t.Fatalf("the tag rule is not reported:\n%s", out)
	}

	// An exception clears the one edge.
	w.write("smith.toml", w.read("smith.toml")+`
[[boundaries.exceptions]]
from = "web"
to = "admin"
reason = "the shared login page, until it moves into a lib"
`)
	if out = w.smith(true, "boundaries"); !strings.Contains(out, "boundaries OK") {
		t.Fatalf("the exception did not clear the edge:\n%s", out)
	}

	// Without the exception, removing the import clears it as well.
	w.write("smith.toml", config+"\n[boundaries]\nkinds = true\n")
	w.write("apps/web/main.go", webImportsNothing)
	if out = w.smith(true, "boundaries"); !strings.Contains(out, "boundaries OK") {
		t.Fatalf("removing the import did not clear the violation:\n%s", out)
	}
}
