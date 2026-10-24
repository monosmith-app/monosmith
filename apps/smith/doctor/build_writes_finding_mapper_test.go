package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestBuildWritesFindingNamesTheGapAndFixesIt(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "packages", "auth"), 0o755); err != nil {
		t.Fatal(err)
	}
	script := `{"scripts": {"build": "esbuild src/index.ts --bundle --outfile=dist/index.esm.js && tsc --declarationDir dist/src"}}`
	if err := os.WriteFile(filepath.Join(dir, "packages", "auth", "package.json"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "smith.toml"), []byte("[workspace]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	task := protocol.Task{Command: []string{"npm", "run", "build", "--silent"}, Outputs: []string{"../../dist/packages/auth"}}
	auth := &projectgraph.Project{Project: protocol.Project{Name: "auth", Root: "packages/auth", Tasks: map[string]protocol.Task{"build": task}}, Plugin: "node", Sources: map[string]string{"build.outputs": "plugin node"}}
	s := &session.Session{Root: dir, Graph: &projectgraph.Graph{Projects: []*projectgraph.Project{auth}}}

	f := buildWritesFinding(s)
	if f.OK || !strings.Contains(f.Detail, "auth:build writes packages/auth/dist but its task caches dist/packages/auth") {
		t.Fatalf("%+v", f)
	}
	if !strings.Contains(f.Remedy, `outputs = ["dist"] under [projects.auth.tasks.build]`) {
		t.Fatalf("remedy %q", f.Remedy)
	}
	if f.Fix == nil || !f.Fix.Safe || !strings.Contains(f.Fix.Files[0].Content, "[projects.auth.tasks.build]\noutputs = [\"dist\"]") {
		t.Fatalf("fix %+v", f.Fix)
	}

	auth.Tasks["build"] = protocol.Task{Command: task.Command, Outputs: []string{"dist"}}
	if f := buildWritesFinding(s); !f.OK || f.Fix != nil {
		t.Fatalf("declared outputs pass: %+v", f)
	}
}

func TestBuildWritesFindingSummarisesManyProjectsButFixesThemAll(t *testing.T) {
	dir := t.TempDir()
	var projects []*projectgraph.Project
	for i := 0; i < 23; i++ {
		name := fmt.Sprintf("pkg%02d", i)
		root := "packages/" + name
		if err := os.MkdirAll(filepath.Join(dir, root), 0o755); err != nil {
			t.Fatal(err)
		}
		script := `{"scripts": {"build": "tsc --outDir dist"}}`
		if err := os.WriteFile(filepath.Join(dir, root, "package.json"), []byte(script), 0o644); err != nil {
			t.Fatal(err)
		}
		task := protocol.Task{Command: []string{"npm", "run", "build", "--silent"}, Outputs: []string{"../../dist/" + root}}
		projects = append(projects, &projectgraph.Project{Project: protocol.Project{Name: name, Root: root, Tasks: map[string]protocol.Task{"build": task}}, Plugin: "node", Sources: map[string]string{"build.outputs": "plugin node"}})
	}
	if err := os.WriteFile(filepath.Join(dir, "smith.toml"), []byte("[workspace]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	f := buildWritesFinding(&session.Session{Root: dir, Graph: &projectgraph.Graph{Projects: projects}})
	if f.OK || strings.Count(f.Detail, "writes") != 3 || !strings.Contains(f.Detail, "and 20 more") {
		t.Fatalf("three spelled out, the rest counted: %s", f.Detail)
	}
	if strings.Count(f.Remedy, "\n") > 4 || !strings.Contains(f.Remedy, "smith doctor --fix") {
		t.Fatalf("not 23 lines: %q", f.Remedy)
	}
	if f.Fix == nil || !f.Fix.Safe || len(f.Fix.AppliesTo) != 23 || strings.Count(f.Fix.Files[0].Content, "[projects.pkg") != 23 {
		t.Fatalf("one safe fix declaring all 23: %+v", f.Fix)
	}
	plan := Plan(Report{Findings: []Finding{{Finding: f}}})
	if len(plan) != 1 || doctorfix.Scope(plan[0].AppliesTo) != "applies to 23 projects: pkg00, pkg01, pkg02, pkg03 and 19 more" {
		t.Fatalf("%+v", plan)
	}
}
