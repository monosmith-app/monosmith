package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const foldSample = `schema = 1

[workspace]
name = "acme"

# libraries build into dist
[projects.a.tasks.build]
outputs = ["dist"]

[projects.b.tasks.build]
outputs = ["dist"]
`

func TestRepeatedOverridesAreSuggestedAsAGroupWithAFix(t *testing.T) {
	cfg, err := workspace.Parse([]byte(foldSample))
	if err != nil {
		t.Fatal(err)
	}
	var graph []*projectgraph.Project
	for _, n := range []string{"a", "b"} {
		graph = append(graph, &projectgraph.Project{Plugin: "node", Sources: map[string]string{"build.command": "plugin node"}, Project: protocol.Project{
			Name: n, Root: "libs/" + n, Language: "node", Kind: protocol.ProjectLib,
			Tasks: map[string]protocol.Task{"build": {Command: []string{"tsc"}}},
		}})
	}
	f, ok := TaskFoldFinding(graph, cfg, foldSample)
	if !ok || !f.Warning || !strings.Contains(f.Detail, "build (2 projects, lang:node and kind:lib)") {
		t.Fatalf("finding = %+v", f)
	}
	if f.Fix == nil || len(f.Fix.Files) != 1 || f.Fix.Files[0].Path != workspace.FileName {
		t.Fatalf("fix = %+v", f.Fix)
	}
	text := f.Fix.Files[0].Content
	if !strings.Contains(text, "[[task_groups]]") || strings.Contains(text, "[projects.a.tasks.build]") || !strings.Contains(text, "[workspace]") {
		t.Errorf("text:\n%s", text)
	}

	if _, ok := TaskFoldFinding(graph[:1], cfg, foldSample); ok {
		t.Error("one project repeats nothing")
	}
}
