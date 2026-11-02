package buildoutputs

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func projectWithOutputs(name, root string, outputs map[string][]string) *projectgraph.Project {
	tasks := map[string]protocol.Task{}
	for task, o := range outputs {
		tasks[task] = protocol.Task{Outputs: o}
	}
	return &projectgraph.Project{Project: protocol.Project{Name: name, Root: root, Tasks: tasks}}
}

func TestSharedOutputsNamesTasksThatWriteTheSameDirectory(t *testing.T) {
	host := projectWithOutputs("azure-host", "apps/azure-host", map[string][]string{"package": {"../../dist/drop"}, "build": {"../../dist/apps/azure-host"}})
	ui := projectWithOutputs("studio-ui", "apps/studio-ui", map[string][]string{"package": {"../../dist/drop"}, "build": {"../../dist/apps/studio-ui"}})
	got := SharedOutputs([]*projectgraph.Project{host, ui})
	if strings.Join(got, "\n") != "azure-host:package and studio-ui:package both write dist/drop" {
		t.Errorf("got:\n%s", strings.Join(got, "\n"))
	}
}

func TestSharedOutputsNamesADirectoryThatHoldsAnother(t *testing.T) {
	a := projectWithOutputs("a", "libs/a", map[string][]string{"package": {"../../dist/drop"}})
	b := projectWithOutputs("b", "libs/b", map[string][]string{"package": {"../../dist/drop/b"}})
	got := SharedOutputs([]*projectgraph.Project{a, b})
	if len(got) != 1 || got[0] != "a:package and b:package both write dist/drop and dist/drop/b" {
		t.Errorf("got %v", got)
	}
}

func TestSharedOutputsLeavesSeparateDirectoriesAlone(t *testing.T) {
	a := projectWithOutputs("a", "libs/a", map[string][]string{"package": {"../../dist/drop/a"}, "build": {"../../dist/libs/a"}})
	b := projectWithOutputs("b", "libs/b", map[string][]string{"package": {"../../dist/drop/b"}, "build": {"../../dist/libs/b"}})
	if got := SharedOutputs([]*projectgraph.Project{a, b}); len(got) != 0 {
		t.Errorf("got %v", got)
	}
	// One task naming a directory twice is not a clash with itself.
	c :=projectWithOutputs("c", "libs/c", map[string][]string{"build": {"dist", "dist/types"}})
	if got := SharedOutputs([]*projectgraph.Project{c}); len(got) != 0 {
		t.Errorf("got %v", got)
	}
}

func TestOverlapWarningsNameBothTasksAndTheFix(t *testing.T) {
	zip := projectWithOutputs("mvd", "apps/mvd", map[string][]string{"package-zip": {"../../dist/drop"}, "package-dmg": {"../../dist/drop"}})
	other := projectWithOutputs("tool", "apps/tool", map[string][]string{"package": {"../../dist/drop"}})
	projects := []*projectgraph.Project{zip, other}
	got := OverlapWarnings(projects)
	if len(got) == 0 {
		t.Fatal("no warning for tasks that write dist/drop")
	}
	for _, want := range []string{"mvd:package-dmg", "mvd:package-zip", "dist/drop", "shared_outputs"} {
		if !strings.Contains(strings.Join(got, "\n"), want) {
			t.Errorf("warnings miss %q:\n%s", want, strings.Join(got, "\n"))
		}
	}
	if got := OverlapWarnings(projects, "dist/drop"); len(got) != 0 {
		t.Errorf("a shared folder is not warned about: %v", got)
	}
}
