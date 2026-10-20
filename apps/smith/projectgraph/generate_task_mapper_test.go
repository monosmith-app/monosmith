package projectgraph

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerateTaskReadsTheSchemaAndWritesTheOutDir(t *testing.T) {
	task, err := GenerateTaskFor("clients/go", workspace.GenerateConfig{
		Schema: "services/orders/openapi.yaml", Generator: "go", Version: "7.8.0",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"npx", "--yes", "@openapitools/openapi-generator-cli@7.8.0", "generate", "-i", "../../services/orders/openapi.yaml", "-g", "go", "-o", "generated"}
	if !reflect.DeepEqual(task.Command, want) {
		t.Fatalf("command %v", task.Command)
	}
	if !task.Cache || !reflect.DeepEqual(task.Inputs, []string{"{workspaceRoot}/services/orders/openapi.yaml"}) || !reflect.DeepEqual(task.Outputs, []string{"generated"}) {
		t.Fatalf("task %+v", task)
	}
	if task.Env["SMITH_GENERATOR_VERSION"] != "7.8.0" {
		t.Fatal("version not in the environment")
	}
}

func TestGenerateTaskProtobufAndCustomCommand(t *testing.T) {
	task, err := GenerateTaskFor("libs/py", workspace.GenerateConfig{Schema: "contracts/user.proto", Generator: "python", Version: "27.0", Out: "src/gen"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"protoc", "-I", "../../contracts", "--python_out=src/gen", "../../contracts/user.proto"}
	if !reflect.DeepEqual(task.Command, want) {
		t.Fatalf("command %v", task.Command)
	}
	custom, err := GenerateTaskFor("libs/py", workspace.GenerateConfig{Schema: "contracts/user.proto", Version: "1", Command: []string{"buf", "generate", "{schema}", "{out}"}})
	if err != nil || custom.Command[2] != "../../contracts/user.proto" {
		t.Fatalf("custom %v %v", custom.Command, err)
	}
}

func TestGenerateTaskRejectsIncompleteConfig(t *testing.T) {
	for name, g := range map[string]workspace.GenerateConfig{
		"no schema":    {Version: "1", Generator: "go"},
		"unpinned":     {Schema: "a/openapi.yaml", Generator: "go"},
		"no generator": {Schema: "a/openapi.yaml", Version: "1"},
		"out escapes":  {Schema: "a/openapi.yaml", Version: "1", Generator: "go", Out: "../x"},
		"no default":   {Schema: "a/schema.graphql", Version: "1", Generator: "go"},
	} {
		if _, err := GenerateTaskFor("app", g); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestGenerateRunsBeforeTheTasksThatReadTheCode(t *testing.T) {
	p := &Project{Project: protocol.Project{Name: "go-client", Root: "clients/go", Tasks: map[string]protocol.Task{
		"build": {Command: []string{"go", "build"}, DependsOn: []string{"^build"}},
		"test":  {Command: []string{"go", "test"}},
	}}, Sources: map[string]string{}}
	err := p.addGenerate(workspace.GenerateConfig{Schema: "services/orders/openapi.yaml", Generator: "go", Version: "7.8.0"})
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Tasks["build"].DependsOn; !reflect.DeepEqual(got, []string{"generate", "^build"}) {
		t.Fatalf("build deps %v", got)
	}
	if got := p.Tasks["test"].DependsOn; !reflect.DeepEqual(got, []string{"generate"}) {
		t.Fatalf("test deps %v", got)
	}
	if _, ok := p.Tasks["lint"]; ok {
		t.Fatal("added a lint task")
	}
	if !strings.Contains(p.Sources["generate.command"], "generate") {
		t.Fatal("source not recorded")
	}
}

func TestGenerateSchemaIsConsumedWithoutConsumesSchemas(t *testing.T) {
	ps := []*Project{proj("orders", "services/orders", nil, nil), proj("go-client", "clients/go", nil, nil)}
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{
		"go-client": {Generate: &workspace.GenerateConfig{Schema: "services/orders/openapi.yaml"}},
	}}
	ResolveEdges(ps)
	ResolveSchemaEdges(ps, []string{"services/orders/openapi.yaml"}, cfg)
	if !reflect.DeepEqual(ps[1].Deps, []string{"orders"}) {
		t.Fatalf("deps %v", ps[1].Deps)
	}
}
