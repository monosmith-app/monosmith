package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestOutsideBuildDir(t *testing.T) {
	web := &projectgraph.Project{Plugin: "node", Project: protocol.Project{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{
		"build": {Outputs: []string{"../../dist/apps/web"}},
		"test":  {Outputs: []string{"coverage"}},
	}}, Sources: map[string]string{"test.outputs": "smith.toml [tasks.test]"}}
	api := &projectgraph.Project{Plugin: "go", Project: protocol.Project{Name: "api", Root: "apps/api", Tasks: map[string]protocol.Task{"build": {Outputs: []string{"dist"}}}}}
	kit := &projectgraph.Project{Plugin: "node", Project: protocol.Project{Name: "kit", Root: "packages/kit", Tasks: map[string]protocol.Task{"build": {Outputs: []string{"dist"}}}},
		Sources: map[string]string{"build.outputs": "smith.toml [projects.kit.tasks.build]"}}
	got := OutsideBuildDir([]*projectgraph.Project{web, api, kit})
	want := []string{
		"web:test writes apps/web/coverage (from smith.toml [tasks.test]), not dist/apps/web/",
		"api:build writes apps/api/dist (from plugin go), not dist/apps/api/",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s", strings.Join(got, "\n"))
	}
}

func TestSharedWorkOutputsRefusesAWorkFolderAnotherProjectShipsFrom(t *testing.T) {
	api := &projectgraph.Project{Project: protocol.Project{Name: "api", Root: "apps/api", Tasks: map[string]protocol.Task{
		"build": {Outputs: []string{"../../dist/apps/api"}, Ship: []string{"../../dist/cargo/debug/api", "../../dist/cargo/debug/api.exe"}},
	}}}
	cli := &projectgraph.Project{Project: protocol.Project{Name: "cli", Root: "apps/cli", Tasks: map[string]protocol.Task{
		"build": {Outputs: []string{"../../dist/cargo"}},
	}}}
	jar := &projectgraph.Project{Project: protocol.Project{Name: "kit", Root: "libs/kit", Tasks: map[string]protocol.Task{
		"build": {Outputs: []string{"../../dist/apps/kit", "../../dist/maven/kit"}, Ship: []string{"../../dist/maven/kit/*.jar"}},
	}}}
	projects := []*projectgraph.Project{api, cli, jar}
	if got := OutsideBuildDir(projects); len(got) != 0 {
		t.Errorf("a work folder under dist/<tool> is a build folder: %v", got)
	}
	got := SharedWorkOutputs(projects)
	if want := "cli:build caches dist/cargo, where api ships from"; strings.Join(got, "\n") != want {
		t.Fatalf("got %v, want %q (kit's own work folder is not shared)", got, want)
	}
	cli.Tasks["build"] = protocol.Task{Outputs: []string{"../../dist/apps/cli"}}
	if got := SharedWorkOutputs(projects); len(got) != 0 {
		t.Errorf("only the shipped folder: %v", got)
	}
}

func TestAnAllowedOutputOutsideDistIsNotReported(t *testing.T) {
	stage := &projectgraph.Project{Project: protocol.Project{Name: "mvd", Root: "apps/mvd", Tasks: map[string]protocol.Task{"stage-web": {Outputs: []string{"appwindow/web"}}}}}
	projects := []*projectgraph.Project{stage}
	if got := OutsideBuildDir(projects); len(got) != 1 {
		t.Fatalf("without the setting it is reported: %v", got)
	}
	if got := OutsideBuildDir(projects, "apps/mvd/appwindow/web"); len(got) != 0 {
		t.Errorf("with the setting it is not: %v", got)
	}
}

func TestAWebAppThatKeepsItsBuildsInTheProjectIsNotReported(t *testing.T) {
	web := &projectgraph.Project{Plugin: "node", Project: protocol.Project{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{
		"build":     {Outputs: []string{"dist/prd"}},
		"build:uat": {Outputs: []string{"dist/uat"}},
	}, Settings: map[string]map[string]any{"vite": {"output": "project"}}}}
	if got := OutsideBuildDir([]*projectgraph.Project{web}); len(got) != 0 {
		t.Errorf("the owner chose dist/<env> in the app: %v", got)
	}
	web.Settings = nil
	if got := OutsideBuildDir([]*projectgraph.Project{web}); len(got) != 2 {
		t.Errorf("without the choice both are reported: %v", got)
	}
}
