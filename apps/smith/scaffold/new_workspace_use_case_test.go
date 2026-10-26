package scaffold

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestNewWorkspaceInAnExistingClone(t *testing.T) {
	dir := t.TempDir()
	g := gitclient.New(dir)
	if err := g.Init(); err != nil {
		t.Skip("git unavailable")
	}
	_, _ = g.Run("config", "user.email", "t@e.com")
	_, _ = g.Run("config", "user.name", "T")
	_, _ = g.Run("config", "commit.gpgsign", "false")
	_, _ = g.Run("remote", "add", "origin", "git@github.com:Acme/acme-repo.git")
	_ = os.WriteFile(filepath.Join(dir, "LICENSE"), []byte("MIT"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("# mine"), 0o644)

	var out bytes.Buffer
	if _, err := NewWorkspace(NewOptions{Name: "acme", Dir: dir, Providers: []string{"github", "azure"}, Version: "1.0.0", Out: &out}); err != nil {
		t.Fatal(err)
	}
	cfg, err := workspace.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Workspace.Scope != "@acme" || cfg.Workspace.ModulePrefix != "github.com/Acme/acme-repo" || len(cfg.CI.Providers) != 2 {
		t.Fatalf("config %+v", cfg.Workspace)
	}
	if body, _ := os.ReadFile(filepath.Join(dir, "README.md")); string(body) != "# mine" {
		t.Fatal("an existing README must be kept")
	}
	for _, f := range []string{".github/workflows/ci.yml", "azure-pipelines.yml", ".gitignore", "LICENSE"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(f))); err != nil {
			t.Errorf("missing %s", f)
		}
	}
	if !strings.Contains(out.String(), "committed") || !g.HasCommits() {
		t.Fatalf("expected a first commit:\n%s", out.String())
	}
	if _, err := NewWorkspace(NewOptions{Name: "acme", Dir: dir, Out: &out}); err == nil {
		t.Fatal("creating over an existing workspace must fail")
	}
}

// isolatedGit runs git with an identity and none of the machine's settings
// (no signing, no hooks from a global config).
func isolatedGit(t *testing.T) {
	t.Helper()
	global := filepath.Join(t.TempDir(), "gitconfig")
	_ = os.WriteFile(global, nil, 0o644)
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_COMMITTER_NAME"} {
		t.Setenv(k, "T")
	}
	for _, k := range []string{"GIT_AUTHOR_EMAIL", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(k, "t@e.com")
	}
}

// repoWithAnUnrelatedChange is a repository with one commit and notes.txt
// modified since.
func repoWithAnUnrelatedChange(t *testing.T) (*gitclient.Git, string) {
	t.Helper()
	isolatedGit(t)
	root := t.TempDir()
	g := gitclient.New(root)
	if err := g.Init(); err != nil {
		t.Skip("git unavailable")
	}
	notes := filepath.Join(root, "notes.txt")
	_ = os.WriteFile(notes, []byte("one"), 0o644)
	if err := g.CommitAll("chore: first"); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(notes, []byte("two"), 0o644)
	head, _ := g.Run("rev-parse", "HEAD")
	return g, head
}

func TestNewWorkspaceInsideARepositoryCommitsNothing(t *testing.T) {
	g, head := repoWithAnUnrelatedChange(t)
	dir := filepath.Join(g.Dir, "tools", "demo")
	var out bytes.Buffer
	if _, err := NewWorkspace(NewOptions{Name: "demo", Dir: dir, Version: "1.0.0", Out: &out}); err != nil {
		t.Fatal(err)
	}
	if now, _ := g.Run("rev-parse", "HEAD"); now != head {
		t.Errorf("HEAD moved from %s to %s:\n%s", head, now, out.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		t.Error("made a nested repository")
	}
	if status, _ := g.Run("status", "--porcelain", "--", "notes.txt"); !strings.HasPrefix(status, " M") {
		t.Errorf("notes.txt status = %q, want modified and unstaged", status)
	}
	if staged, _ := g.Run("diff", "--cached", "--name-only"); staged != "" {
		t.Errorf("staged %q", staged)
	}
	if !strings.Contains(out.String(), "not committed") || !strings.Contains(out.String(), "inside the git repository") {
		t.Errorf("output does not say it left the workspace uncommitted:\n%s", out.String())
	}
}

func TestNewWorkspaceAtTheTopOfARepositoryCommitsOnlyItsFiles(t *testing.T) {
	g, head := repoWithAnUnrelatedChange(t)
	var out bytes.Buffer
	if _, err := NewWorkspace(NewOptions{Name: "acme", Dir: g.Dir, Version: "1.0.0", Out: &out}); err != nil {
		t.Fatal(err)
	}
	if parent, _ := g.Run("rev-parse", "HEAD~1"); parent != head {
		t.Fatalf("expected one commit on top of %s:\n%s", head, out.String())
	}
	files, _ := g.Run("show", "--name-only", "--format=", "HEAD")
	if strings.Contains(files, "notes.txt") || !strings.Contains(files, workspace.FileName) {
		t.Errorf("commit holds %q", files)
	}
	if status, _ := g.Run("status", "--porcelain", "--", "notes.txt"); !strings.HasPrefix(status, " M") {
		t.Errorf("notes.txt status = %q, want still modified", status)
	}
}

func TestNewWorkspaceOutsideARepositoryCreatesAndCommitsOne(t *testing.T) {
	isolatedGit(t)
	parent := t.TempDir()
	// The temporary directory must not be inside a repository itself.
	t.Setenv("GIT_CEILING_DIRECTORIES", parent)
	dir := filepath.Join(parent, "acme")
	var out bytes.Buffer
	if _, err := NewWorkspace(NewOptions{Name: "acme", Dir: dir, Version: "1.0.0", Out: &out}); err != nil {
		t.Fatal(err)
	}
	g := gitclient.New(dir)
	if top, _ := g.Prefix(); top != "" || !g.HasCommits() {
		t.Fatalf("expected a new repository with a first commit:\n%s", out.String())
	}
	if status, _ := g.Run("status", "--porcelain"); status != "" {
		t.Errorf("left %q uncommitted", status)
	}
}

func TestValidateName(t *testing.T) {
	for _, ok := range []string{"web", "web-app", "a1"} {
		if ValidateName("app", ok) != nil {
			t.Errorf("%s must be valid", ok)
		}
	}
	for _, bad := range []string{"Web", "web_app", "1web", "web--app", "-web", "", "web/app"} {
		if ValidateName("app", bad) == nil {
			t.Errorf("%q must be invalid", bad)
		}
	}
}

func TestSplitKind(t *testing.T) {
	for in, want := range map[string][2]string{"node-app": {"app", "node"}, "python-package": {"package", "python"}, "my-lang-lib": {"lib", "my-lang"}} {
		if k, l, ok := SplitKind(in); !ok || k != want[0] || l != want[1] {
			t.Errorf("SplitKind(%q) = %s %s %v", in, k, l, ok)
		}
	}
	for _, in := range []string{"app", "node-web", "-app"} {
		if _, _, ok := SplitKind(in); ok {
			t.Errorf("SplitKind(%q) must not split", in)
		}
	}
}

func TestAWorkspaceWithoutARemoteHasNoModulePrefixToDerive(t *testing.T) {
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	git := gitclient.New(dir)
	if got := remoteModulePrefix(git); got != "" {
		t.Fatalf("no remote, no prefix: %q", got)
	}
	if got := modulePrefixFromRemote(git, "acme"); got != "example.com/acme" {
		t.Fatalf("new workspaces still fall back: %q", got)
	}
	if out, err := exec.Command("git", "-C", dir, "remote", "add", "origin", "git@github.com:Acme/Tools.git").CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if got := remoteModulePrefix(git); got != "github.com/Acme/Tools" {
		t.Fatalf("from the remote: %q", got)
	}
}
