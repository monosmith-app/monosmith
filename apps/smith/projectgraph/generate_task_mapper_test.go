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
		"no kind":      {Schema: "a/data.json", Version: "1", Generator: "go"},
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
	err := p.addGenerate(workspace.ProjectOverride{Generate: &workspace.GenerateConfig{Schema: "services/orders/openapi.yaml", Generator: "go", Version: "7.8.0"}})
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

func TestGenerateAsyncAPIAndGraphQLDefaults(t *testing.T) {
	async, err := GenerateTaskFor("apps/events", workspace.GenerateConfig{Schema: "specs/asyncapi.yaml", Generator: "@asyncapi/html-template", Version: "6.2.1", Out: "docs"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"npx", "--yes", "@asyncapi/cli@6.2.1", "generate", "fromTemplate", "../../specs/asyncapi.yaml", "@asyncapi/html-template", "-o", "docs", "--force-write"}
	if !reflect.DeepEqual(async.Command, want) {
		t.Fatalf("asyncapi command %v", async.Command)
	}
	for _, generator := range []string{"", "typescript"} {
		gql, err := GenerateTaskFor("apps/web", workspace.GenerateConfig{Schema: "api/schema.graphql", Generator: generator, Version: "6.3.4"})
		if err != nil {
			t.Fatal(err)
		}
		want = []string{"npx", "--yes", "@genql/cli@6.3.4", "--schema", "../../api/schema.graphql", "--output", "generated"}
		if !reflect.DeepEqual(gql.Command, want) {
			t.Fatalf("graphql command %v", gql.Command)
		}
	}
	if _, err := GenerateTaskFor("apps/events", workspace.GenerateConfig{Schema: "specs/asyncapi.yaml", Version: "1"}); err == nil {
		t.Error("an AsyncAPI template is required")
	}
}

func TestSeveralSchemasMakeOneTaskEach(t *testing.T) {
	p := &Project{Project: protocol.Project{Name: "client", Root: "clients/py", Tasks: map[string]protocol.Task{
		"test": {Command: []string{"pytest"}},
	}}, Sources: map[string]string{}}
	err := p.addGenerate(workspace.ProjectOverride{
		Generate: &workspace.GenerateConfig{Schema: "contracts/user.proto", Generator: "python", Version: "36.2"},
		Generates: []workspace.GenerateConfig{
			{Schema: "contracts/order.proto", Generator: "python", Version: "36.2"},
			{Name: "Billing API", Schema: "services/billing/openapi.yaml", Generator: "python", Version: "7.8.0"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Tasks["generate-order"].Outputs; !reflect.DeepEqual(got, []string{"generated-order"}) {
		t.Fatalf("outputs %v", got)
	}
	if _, ok := p.Tasks["generate-billing-api"]; !ok {
		t.Fatalf("tasks %v", p.Tasks)
	}
	if got := p.Tasks["test"].DependsOn; !reflect.DeepEqual(got, []string{"generate", "generate-billing-api", "generate-order"}) {
		t.Fatalf("test deps %v", got)
	}
	// Two entries must not write into each other's directory.
	err = (&Project{Project: protocol.Project{Name: "c", Root: "c"}, Sources: map[string]string{}}).addGenerate(workspace.ProjectOverride{
		Generate:  &workspace.GenerateConfig{Schema: "a/a.proto", Generator: "go", Version: "1"},
		Generates: []workspace.GenerateConfig{{Schema: "a/b.proto", Generator: "go", Version: "1", Out: "generated/b"}},
	})
	if err == nil || !strings.Contains(err.Error(), "overlapping") {
		t.Fatalf("overlap accepted: %v", err)
	}
	err = (&Project{Project: protocol.Project{Name: "c", Root: "c"}, Sources: map[string]string{}}).addGenerate(workspace.ProjectOverride{
		Generates: []workspace.GenerateConfig{{Schema: "a/x.proto", Generator: "go", Version: "1"}, {Schema: "b/x.proto", Generator: "go", Version: "1"}},
	})
	if err == nil || !strings.Contains(err.Error(), "named") {
		t.Fatalf("duplicate name accepted: %v", err)
	}
}

func TestAPluginsGenerateTaskIsOrderedAndConsumesItsSchema(t *testing.T) {
	ps := []*Project{proj("orders", "services/orders", nil, nil), proj("client", "clients/go", nil, nil)}
	client := ps[1]
	client.Tasks = map[string]protocol.Task{"build": {Command: []string{"go", "build"}}}
	merge(client, protocol.Contribution{Tasks: map[string]protocol.Task{
		"generate-orders": {Command: []string{"gen"}, Inputs: []string{"{workspaceRoot}/services/orders/openapi.yaml"}, Outputs: []string{"gen"}, Cache: true},
	}}, "plugin gen")
	ResolveEdges(ps)
	ResolveSchemaEdges(ps, []string{"services/orders/openapi.yaml"}, workspace.Config{})
	if !reflect.DeepEqual(client.Deps, []string{"orders"}) {
		t.Fatalf("deps %v", client.Deps)
	}
	if err := ApplyOverrides(client, workspace.Config{}); err != nil {
		t.Fatal(err)
	}
	if got := client.Tasks["build"].DependsOn; !reflect.DeepEqual(got, []string{"generate-orders"}) {
		t.Fatalf("build deps %v", got)
	}
}
