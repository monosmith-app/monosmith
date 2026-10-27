package projectgraph

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func nodeLib(name string) *Project {
	return &Project{Plugin: "node", Project: protocol.Project{Name: name, Root: "libs/" + name, Language: "node", Kind: protocol.ProjectLib, Tasks: map[string]protocol.Task{
		"build": {Command: []string{"tsc"}},
		"test":  {Command: []string{"jest"}},
	}}}
}

func TestAProjectOverridesItsTaskGroupField(t *testing.T) {
	cfg := workspace.Config{
		Tasks: map[string]workspace.TaskOverride{"build": {Outputs: []string{"out"}, Inputs: []string{"src/**"}}},
		TaskGroups: []workspace.TaskGroup{
			{Select: []string{"lang:node", "kind:lib"}, Tasks: map[string]workspace.TaskOverride{
				"build": {Outputs: []string{"dist"}},
				"test":  {Env: map[string]string{"A": "1"}, DependsOn: []string{"^build"}},
			}},
			{Select: []string{"libs/*"}, Tasks: map[string]workspace.TaskOverride{
				"test": {DependsOn: []string{"^build", "logger:build"}},
			}},
			{Select: []string{"lang:go"}, Tasks: map[string]workspace.TaskOverride{"build": {Outputs: []string{"bin"}}}},
		},
		Projects: map[string]workspace.ProjectOverride{"special": {Tasks: map[string]workspace.TaskOverride{
			"test": {Env: map[string]string{"A": "2"}},
		}}},
	}
	plain, special := nodeLib("plain"), nodeLib("special")
	for _, p := range []*Project{plain, special} {
		if err := ApplyOverrides(p, cfg); err != nil {
			t.Fatal(err)
		}
	}
	if got := plain.Tasks["build"].Outputs; strings.Join(got, ",") != "dist" {
		t.Errorf("group beats [tasks.build]: outputs = %v", got)
	}
	if got := plain.Tasks["build"].Inputs; strings.Join(got, ",") != "src/**" {
		t.Errorf("a field the group leaves alone keeps [tasks.build]: inputs = %v", got)
	}
	if got := plain.Tasks["test"].DependsOn; strings.Join(got, ",") != "^build,logger:build" {
		t.Errorf("a later group wins field by field: depends_on = %v", got)
	}
	if got := plain.Tasks["test"].Env["A"]; got != "1" {
		t.Errorf("env = %q", got)
	}
	if got := special.Tasks["test"].Env["A"]; got != "2" {
		t.Errorf("the project's own override beats the group: env = %q", got)
	}
	if got := special.Tasks["test"].DependsOn; strings.Join(got, ",") != "^build,logger:build" {
		t.Errorf("a project override of one field keeps the group's others: %v", got)
	}
	want := `smith.toml [[task_groups]] select = ["lang:node", "kind:lib"]`
	if got := plain.Sources["build.outputs"]; got != want {
		t.Errorf("source = %q, want %q", got, want)
	}
	if got := plain.Sources["test.dependsOn"]; got != `smith.toml [[task_groups]] select = ["libs/*"]` {
		t.Errorf("source = %q", got)
	}
	if got := special.Sources["test.env"]; got != "smith.toml [projects.special.tasks.test]" {
		t.Errorf("source = %q", got)
	}
}

func TestAGroupSeesTagsFromSmithToml(t *testing.T) {
	p := nodeLib("a")
	cfg := workspace.Config{
		TaskGroups: []workspace.TaskGroup{{Select: []string{"tag:ui"}, Tasks: map[string]workspace.TaskOverride{"build": {Outputs: []string{"dist"}}}}},
		Projects:   map[string]workspace.ProjectOverride{"a": {Tags: []string{"ui"}}},
	}
	if err := ApplyOverrides(p, cfg); err != nil {
		t.Fatal(err)
	}
	if len(p.Tasks["build"].Outputs) != 1 {
		t.Errorf("outputs = %v", p.Tasks["build"].Outputs)
	}
}
