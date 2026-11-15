package explain

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestExplainShowsSources(t *testing.T) {
	p := &projectgraph.Project{Project: protocol.Project{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{
		"build": {Command: []string{"npm", "run", "build"}, Outputs: []string{"dist"}, Cache: true},
		"test":  {Command: []string{"npm", "test"}},
	}}, Plugin: "node"}
	off := false
	if err := projectgraph.ApplyOverrides(p, workspace.Config{Tasks: map[string]workspace.TaskOverride{"build": {Cache: &off}}}); err != nil {
		t.Fatal(err)
	}
	g, _ := projectgraph.New([]*projectgraph.Project{p})
	tasks, err := Explain(g, "web:build")
	if err != nil || len(tasks) != 1 {
		t.Fatalf("%v %v", tasks, err)
	}
	fields := map[string]Field{}
	for _, f := range tasks[0].Fields {
		fields[f.Name] = f
	}
	if fields["cache"].Value != "false" || fields["cache"].Source != "smith.toml [tasks.build]" || fields["command"].Source != "plugin node" {
		t.Fatalf("fields %+v", fields)
	}
	if all, _ := Explain(g, "web"); len(all) != 2 {
		t.Fatal("explaining a project lists every task")
	}
	if _, err := Explain(g, "web:nope"); err == nil {
		t.Fatal("unknown task")
	}
	if _, err := Explain(g, "nope"); err == nil {
		t.Fatal("unknown project")
	}
}
