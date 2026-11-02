package releaseassets

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// repoWithTwoCommits makes a scratch repository: commit one holds smith.toml
// and tag app@1.0.0, commit two adds a file. It returns the repository, the
// tag's commit and the HEAD commit.
func repoWithTwoCommits(t *testing.T) (g *gitclient.Git, tagged, head string) {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q")
	if err := os.WriteFile(filepath.Join(dir, "smith.toml"), []byte("[workspace]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "one")
	run("tag", "app@1.0.0")
	tagged = run("rev-parse", "HEAD")
	if err := os.WriteFile(filepath.Join(dir, "later.txt"), []byte("later"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "two")
	return gitclient.New(dir), tagged, run("rev-parse", "HEAD")
}

func TestWithTagWorktreeBuildsTheTagsOwnCommitAndCleansUp(t *testing.T) {
	g, tagged, head := repoWithTwoCommits(t)
	var seen string
	err := withTagWorktree(g, "app@1.0.0", func(dir string) error {
		seen = dir
		at, err := gitclient.New(dir).RevParse("HEAD")
		if err != nil || at != tagged || at == head {
			t.Errorf("the worktree is at %s (%v), want the tag's commit %s", at, err, tagged)
		}
		if _, err := os.Stat(filepath.Join(dir, "later.txt")); err == nil {
			t.Error("the worktree holds a file of a later commit")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(seen); err == nil {
		t.Errorf("the worktree %s was not removed", seen)
	}
	if list, _ := g.Run("worktree", "list", "--porcelain"); strings.Count(list, "worktree ") != 1 {
		t.Errorf("git still lists the worktree:\n%s", list)
	}
}

func TestWithTagWorktreeCleansUpWhenTheBuildFails(t *testing.T) {
	g, _, _ := repoWithTwoCommits(t)
	boom := errors.New("build failed")
	var seen string
	if err := withTagWorktree(g, "app@1.0.0", func(dir string) error { seen = dir; return boom }); !errors.Is(err, boom) {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(seen); err == nil {
		t.Errorf("the worktree %s was not removed", seen)
	}
}

func TestWithTagWorktreeRefusesAnUnknownTag(t *testing.T) {
	g, _, _ := repoWithTwoCommits(t)
	if err := withTagWorktree(g, "app@9.9.9", func(string) error { t.Fatal("must not run"); return nil }); err == nil || !strings.Contains(err.Error(), "app@9.9.9") {
		t.Errorf("got %v", err)
	}
}

func TestHasConfigAtTellsACommitWithoutSmithToml(t *testing.T) {
	g, tagged, _ := repoWithTwoCommits(t)
	if !hasConfigAt(g, tagged) {
		t.Error("the tagged commit has smith.toml")
	}
	// The empty tree stands for a commit from before smith.toml existed.
	if hasConfigAt(g, "4b825dc642cb6eb9a060e54bf8d69288fbee4904") {
		t.Error("a tree without smith.toml has no config")
	}
}
