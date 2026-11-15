package gitclient

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// NewTestRepo creates a repository with an identity, for tests.
func newTestRepo(t *testing.T) *Git {
	t.Helper()
	g := New(t.TempDir())
	if err := g.Init(); err != nil {
		t.Skip("git is not available: ", err)
	}
	_, _ = g.Run("config", "user.email", "test@example.com")
	_, _ = g.Run("config", "user.name", "Test")
	_, _ = g.Run("config", "commit.gpgsign", "false")
	_, _ = g.Run("config", "tag.gpgsign", "false")
	return g
}

func write(t *testing.T, g *Git, rel, body string) {
	t.Helper()
	p := filepath.Join(g.Dir, filepath.FromSlash(rel))
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFilesChangesLogAndTags(t *testing.T) {
	g := newTestRepo(t)
	write(t, g, ".gitignore", "dist/\n")
	write(t, g, "apps/web/a b.txt", "1")
	write(t, g, "apps/web/dist/out.js", "ignored")
	if err := g.CommitAll("feat(web): first\n\nbody line"); err != nil {
		t.Fatal(err)
	}
	files, err := g.Files()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(files, []string{".gitignore", "apps/web/a b.txt"}) {
		t.Fatalf("files %v", files)
	}
	// A tracked file deleted but not committed is not part of the workspace.
	_ = os.Remove(filepath.Join(g.Dir, "apps", "web", "a b.txt"))
	if files, _ := g.Files(); !reflect.DeepEqual(files, []string{".gitignore"}) {
		t.Fatalf("after deleting a b.txt: %v", files)
	}
	write(t, g, "apps/web/a b.txt", "1")
	base, _ := g.RevParse("HEAD")
	if err := g.CreateTag("web@0.1.0", "web 0.1.0"); err != nil {
		t.Fatal(err)
	}
	write(t, g, "libs/ui/x.txt", "2")
	_ = g.CommitAll("fix(ui): second")
	write(t, g, "untracked.txt", "u")
	write(t, g, "apps/web/a b.txt", "changed")

	changed, err := g.ChangedSince(base)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(changed, []string{"apps/web/a b.txt", "libs/ui/x.txt", "untracked.txt"}) {
		t.Fatalf("changed %v", changed)
	}
	log, err := g.Log("web@0.1.0", "libs/ui")
	if err != nil || len(log) != 1 || log[0].Subject != "fix(ui): second" {
		t.Fatalf("log %+v %v", log, err)
	}
	all, _ := g.Log("", "")
	if len(all) != 2 || all[1].Body != "body line" {
		t.Fatalf("all %+v", all)
	}
	tags, _ := g.Tags("web@")
	if !reflect.DeepEqual(tags, []string{"web@0.1.0"}) || g.TagCommit("web@0.1.0") != base || g.TagCommit("nope") != "" {
		t.Fatalf("tags %v", tags)
	}
	if g.CurrentBranch() != "main" || g.IsShallow() || !g.HasCommits() {
		t.Fatal("repository state")
	}
	if err := g.RestoreFiles("apps/web/a b.txt"); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(g.Dir, "apps", "web", "a b.txt"))
	if string(body) != "1" {
		t.Fatalf("restore: %q", body)
	}
}
