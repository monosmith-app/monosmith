package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
