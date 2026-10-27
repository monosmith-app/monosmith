//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDoctorKindHints: a lib no project uses and a package that is never
// published are suggestions pointing at smith kind, once the lib has been in
// the repository for a while, never in a fresh workspace, and never failures.
func TestDoctorKindHints(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "kh", "--ci", "github")
	w.dir = filepath.Join(base, "kh")
	w.smith(true, "add", "lib", "core", "--lang", "go")
	w.smith(true, "add", "app", "cli", "--lang", "go")
	w.smith(true, "add", "package", "sdk", "--lang", "go")
	w.write("smith.toml", w.read("smith.toml")+"\n[projects.sdk]\nregistry = \"none\"\n")
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: a workspace")

	fresh := w.smith(true, "doctor")
	if strings.Contains(fresh, "the lib core is used by a project") {
		t.Fatalf("a fresh workspace is not nagged about its lib:\n%s", fresh)
	}
	if !strings.Contains(fresh, "the package sdk is published") {
		t.Fatalf("a package that its own settings never publish is named at once:\n%s", fresh)
	}

	old := time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
	w.env = append(w.env, "GIT_AUTHOR_DATE="+old, "GIT_COMMITTER_DATE="+old)
	w.git("commit", "--amend", "--no-edit", "-q")
	settled := w.smith(true, "doctor")
	for _, want := range []string{"the lib core is used by a project", "smith kind core package", "smith kind sdk lib"} {
		if !strings.Contains(settled, want) {
			t.Fatalf("doctor lacks %q:\n%s", want, settled)
		}
	}
	if strings.Contains(settled, "✖") {
		t.Fatalf("suggestions are warnings, never failures:\n%s", settled)
	}
	if !strings.Contains(w.smith(true, "doctor", "--json"), `"failed": 0`) {
		t.Fatal("the failed count stays at zero")
	}

	// Following the suggestion makes it go away.
	w.smith(true, "kind", "sdk", "lib")
	if after := w.smith(true, "doctor"); strings.Contains(after, "the package sdk is published") {
		t.Fatalf("a lib is not suggested to be published:\n%s", after)
	}
}
