package cipipeline

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func project(name string, tags []string, tasks map[string]protocol.Task) *projectgraph.Project {
	return &projectgraph.Project{Project: protocol.Project{Name: name, Tags: tags, Tasks: tasks}}
}

func TestMarkedForVerifyListsTheMarkedTasksThatCanRun(t *testing.T) {
	image := protocol.Task{Command: []string{"docker", "build", "."}, Verify: true}
	projects := []*projectgraph.Project{
		project("api", nil, map[string]protocol.Task{"image": image, "build": {Command: []string{"go", "build"}, Verify: true}}),
		project("web", nil, map[string]protocol.Task{"image": {Command: []string{"docker", "build", "."}}}),
		project("driver", []string{release.NativeTag}, map[string]protocol.Task{"image": image}),
		project("host", nil, map[string]protocol.Task{"image": image, "check": {Command: []string{"{smith}", "x"}, Verify: true}}),
	}
	byTask, missing := MarkedForVerify(projects, func(string) bool { return true })
	want := map[string][]string{"image": {"api", "host"}, "check": {"host"}}
	if !reflect.DeepEqual(byTask, want) || len(missing) != 0 {
		t.Errorf("byTask %v, missing %v", byTask, missing)
	}
	byTask, missing = MarkedForVerify(projects, func(p string) bool { return p != "docker" })
	if len(byTask["image"]) != 0 || len(missing) != 2 || !strings.Contains(missing[0], "api:image not verified: docker is not installed") {
		t.Errorf("without docker: byTask %v, missing %v", byTask, missing)
	}
}

// rewritingWorkspace is a git repository whose project core has a build
// task that rewrites libs/core/state.txt, as tsc rewrites a .tsbuildinfo.
func rewritingWorkspace(t *testing.T) (*session.Session, *gitclient.Git) {
	t.Helper()
	for _, name := range []string{"CI", "SMITH_BASE", "GITHUB_BASE_REF", "SYSTEM_PULLREQUEST_TARGETBRANCH", "CI_MERGE_REQUEST_TARGET_BRANCH_NAME", "GITHUB_STEP_SUMMARY"} {
		t.Setenv(name, "")
	}
	root := t.TempDir()
	g := gitclient.New(root)
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "t@example.com"}, {"config", "user.name", "t"}, {"config", "commit.gpgsign", "false"}} {
		if _, err := g.Run(args...); err != nil {
			t.Skip("git unavailable: ", err)
		}
	}
	state := filepath.Join(root, "libs", "core", "state.txt")
	_ = os.MkdirAll(filepath.Dir(state), 0o755)
	_ = os.WriteFile(state, []byte("[a]\n\tb = 0\n"), 0o644)
	if _, err := g.Run("add", "-A"); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Run("commit", "-q", "-m", "chore: start"); err != nil {
		t.Fatal(err)
	}
	// git config --file writes the file: a task rewriting a tracked file on
	// any OS, with nothing to install.
	build := protocol.Task{Command: []string{"git", "config", "--file", "state.txt", "a.b", "1"}}
	core := &projectgraph.Project{Project: protocol.Project{Name: "core", Root: "libs/core", Tasks: map[string]protocol.Task{"build": build}}, Plugin: "test"}
	pg, err := projectgraph.New([]*projectgraph.Project{core})
	if err != nil {
		t.Fatal(err)
	}
	return &session.Session{Root: root, Git: g, Graph: pg, Files: []string{"libs/core/state.txt"}}, g
}

func TestVerifyFailsWhenATaskRewritesATrackedFile(t *testing.T) {
	s, _ := rewritingWorkspace(t)
	var out bytes.Buffer
	err := RunPhase(context.Background(), s, "verify", &out)
	if err == nil {
		t.Fatalf("verify passed with a rewritten tracked file:\n%s", out.String())
	}
	if !strings.Contains(err.Error(), "libs/core/state.txt (in core, whose build ran)") || !strings.Contains(err.Error(), "git rm --cached") {
		t.Errorf("error does not name the file, the task and the fix:\n%v", err)
	}
}

func TestVerifyPassesWhenTheRewrittenFileIsNotTracked(t *testing.T) {
	s, g := rewritingWorkspace(t)
	_ = os.WriteFile(filepath.Join(s.Root, ".gitignore"), []byte("state.txt\n"), 0o644)
	for _, args := range [][]string{{"rm", "-q", "--cached", "libs/core/state.txt"}, {"add", ".gitignore"}, {"commit", "-q", "-m", "chore: stop tracking build state"}} {
		if _, err := g.Run(args...); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	if err := RunPhase(context.Background(), s, "verify", &out); err != nil {
		t.Fatalf("verify failed on an ignored, untracked file: %v\n%s", err, out.String())
	}
}

func TestVerifyLeavesADevelopersOwnEditsAlone(t *testing.T) {
	s, _ := rewritingWorkspace(t)
	// The developer's edit, before verify: not verify's doing on a laptop.
	readme := filepath.Join(s.Root, "README.md")
	_ = os.WriteFile(readme, []byte("x\n"), 0o644)
	if _, err := s.Git.Run("add", "README.md"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Git.Run("commit", "-q", "-m", "docs: readme"); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(readme, []byte("edited\n"), 0o644)
	s.Graph.Projects[0].Tasks = map[string]protocol.Task{"build": {Command: []string{"git", "--version"}}}
	var out bytes.Buffer
	if err := RunPhase(context.Background(), s, "verify", &out); err != nil {
		t.Fatalf("verify failed on the developer's own edit: %v", err)
	}
	t.Setenv("CI", "true")
	if err := RunPhase(context.Background(), s, "verify", &out); err == nil || !strings.Contains(err.Error(), "README.md (already changed before verify started") {
		t.Errorf("in CI a file changed before verify fails it: %v", err)
	}
}
