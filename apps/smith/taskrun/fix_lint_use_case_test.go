package taskrun

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// fixFixture is a workspace of three projects: "left" whose plain lint fails
// even after the fix, "clean" whose plain lint passes after the fix, and
// "plain" whose linter has no fix mode.
func fixFixture(t *testing.T) (root string, pg *projectgraph.Graph, files []string, store taskcache.Storage) {
	t.Helper()
	root = t.TempDir()
	self, _ := os.Executable()
	cmd := func(arg string) []string { return []string{self, "-test.run=^$", arg} }
	env := map[string]string{"SMITH_HELPER": "1"}
	var projects []*projectgraph.Project
	for _, p := range []struct {
		name string
		lint protocol.Task
	}{
		{"left", protocol.Task{Command: cmd("fail"), Fix: &protocol.TaskFix{Command: cmd("fixed")}}},
		{"clean", protocol.Task{Command: cmd("clean"), Fix: &protocol.TaskFix{Command: cmd("fixed")}}},
		{"plain", protocol.Task{Command: cmd("clean")}},
	} {
		dir := filepath.Join(root, "libs", p.name)
		_ = os.MkdirAll(dir, 0o755)
		_ = os.WriteFile(filepath.Join(dir, "src.txt"), []byte("v1"), 0o644)
		files = append(files, "libs/"+p.name+"/src.txt")
		p.lint.Env, p.lint.Cache, p.lint.Outputs = env, true, []string{"out"}
		if p.lint.Fix != nil {
			f := *p.lint.Fix
			p.lint.Fix = &f
		}
		projects = append(projects, &projectgraph.Project{Project: protocol.Project{Name: p.name, Root: "libs/" + p.name, Tasks: map[string]protocol.Task{"lint": p.lint}}, Plugin: "test"})
	}
	pg, err := projectgraph.New(projects)
	if err != nil {
		t.Fatal(err)
	}
	store, _ = taskcache.NewLocalStorage(filepath.Join(root, ".cache"))
	return root, pg, files, store
}

func runLint(t *testing.T, root string, pg *projectgraph.Graph, files []string, store taskcache.Storage, fix bool) (Summary, string) {
	t.Helper()
	g, err := taskgraph.Build(pg, []string{"clean", "left", "plain"}, []string{"lint"})
	if err != nil {
		t.Fatal(err)
	}
	if fix {
		g.UseFix()
	}
	var out bytes.Buffer
	s := Run(context.Background(), g, Options{Root: root, Files: files, Projects: pg, Storage: store, Parallel: 2, Out: &out, ToolVersion: "test", Fix: fix})
	return s, out.String()
}

func TestAFixingLintRunFixesJudgesByThePlainLintAndNeverUsesTheCache(t *testing.T) {
	root, pg, files, store := fixFixture(t)
	// A plain run first: it stores a result for every lint task.
	plain, _ := runLint(t, root, pg, files, store, false)
	if plain.Count(Succeeded) != 2 || plain.Count(Failed) != 1 {
		t.Fatalf("plain run: %s", plain)
	}
	_ = os.RemoveAll(filepath.Join(root, "libs", "clean", "out"))

	s, out := runLint(t, root, pg, files, store, true)
	if s.Count(Cached) != 0 {
		t.Fatalf("a fixing run must not read the cache: %s\n%s", s, out)
	}
	if body, _ := os.ReadFile(filepath.Join(root, "libs", "left", "out", "result.txt")); string(body) != "fixed" {
		t.Fatalf("the fix did not run in left: %q", body)
	}
	if s.OK() {
		t.Fatalf("a problem is left in left: the run must fail\n%s", s)
	}
	text := s.String()
	for _, want := range []string{
		"left  ",
		"▲ problems left",
		"clean",
		"✓ no problems left",
		"▲ No fix mode for plain: the plain lint ran, nothing was changed",
		"▲ Lint fix found problems in ",
		"→ see them: smith run lint -p left",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("report lacks %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "tasks succeeded") {
		t.Fatalf("a run of lint tasks closes with the fix report only:\n%s", text)
	}
}

func TestAFixingLintRunStoresNothingInTheCache(t *testing.T) {
	root, pg, files, store := fixFixture(t)
	s, _ := runLint(t, root, pg, files, store, true)
	for _, r := range s.Results {
		if ok, _ := store.Has(taskcache.AC, r.Key); ok {
			t.Fatalf("%s was cached by a fixing run", r.Task)
		}
	}
}
