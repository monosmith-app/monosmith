//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAddThenUndoLeavesTheWorkspaceAsItWas: with commit_operations on, smith add
// prints its plan and makes one commit; smith undo shows what it reverts and
// makes a revert commit, and the tree of HEAD is the tree before the add.
func TestAddThenUndoLeavesTheWorkspaceAsItWas(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "ops", "--ci", "github", "--default-stack", "node=node-vitest")
	w.dir = filepath.Join(base, "ops")
	toml := strings.Replace(w.read("smith.toml"), "[workspace]", "[workspace]\ncommit_operations = true", 1)
	w.write("smith.toml", toml)
	w.git("add", "-A")
	w.git("commit", "-m", "chore: commit operations")
	original := w.git("rev-parse", "HEAD^{tree}")
	commits := strings.Count(w.git("log", "--format=%H"), "\n")

	out := w.smith(true, "add", "package", "billing", "--lang", "node")
	for _, want := range []string{"Plan: add package billing", "packages/billing/", "Done: package billing is ready", "chore: add package billing"} {
		if !strings.Contains(out, want) {
			t.Errorf("the add output lacks %q:\n%s", want, out)
		}
	}
	if got := strings.Count(w.git("log", "--format=%H"), "\n"); got != commits+1 {
		t.Fatalf("add made %d commits, want 1", got-commits)
	}
	if status := w.git("status", "--porcelain"); status != "" {
		t.Fatalf("the work tree is not clean after add:\n%s", status)
	}

	dry := w.smith(true, "undo", "--dry-run")
	if !strings.Contains(dry, "remove  packages/billing/package.json") || !strings.Contains(dry, "nothing was changed") {
		t.Fatalf("undo --dry-run does not show what it reverts:\n%s", dry)
	}
	w.smith(true, "undo", "--yes")
	if tree := w.git("rev-parse", "HEAD^{tree}"); tree != original {
		t.Fatalf("the tree after undo differs from the original:\n%s", w.git("diff", "--stat", original, "HEAD"))
	}
	if subject := strings.TrimSpace(w.git("log", "-1", "--format=%s")); subject != "chore: undo add package billing" {
		t.Fatalf("subject: %q", subject)
	}
	w.smith(false, "undo", "--yes")
}
