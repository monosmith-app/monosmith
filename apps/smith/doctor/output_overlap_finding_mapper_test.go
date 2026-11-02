package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ownerProjects are the tasks of the report: api:build and api:build-dev write
// apps/api/dist; models:build and models:watch write libs/models/dist.
func ownerProjects() []*projectgraph.Project {
	api := &projectgraph.Project{Project: protocol.Project{Name: "api", Root: "apps/api", Tasks: map[string]protocol.Task{
		"build":     {Command: []string{"npm", "run", "build", "--silent"}, Outputs: []string{"dist"}, Cache: true},
		"build-dev": {Command: []string{"npm", "run", "build-dev", "--silent"}, Outputs: []string{"dist"}, Cache: true},
	}}}
	models := &projectgraph.Project{Project: protocol.Project{Name: "models", Root: "libs/models", Tasks: map[string]protocol.Task{
		"build": {Outputs: []string{"dist"}, Cache: true},
		"watch": {Outputs: []string{"dist"}},
	}}}
	return []*projectgraph.Project{api, models}
}

func TestOverlapFindingNamesTheTasksAndSaysWhatGoesWrong(t *testing.T) {
	f := outputOverlapFinding(ownerProjects(), nil, nil, "[workspace]\nname = \"w\"\n")
	if f.OK || !f.Warning {
		t.Fatalf("want a warning, got %+v", f)
	}
	for _, want := range []string{"api:build and api:build-dev both write apps/api/dist", "clears its output folder", "deletes the files the other task wrote"} {
		if !strings.Contains(f.Detail, want) {
			t.Errorf("detail misses %q:\n%s", want, f.Detail)
		}
	}
}

func TestOverlapFindingLeavesAWatchTaskAlone(t *testing.T) {
	f := outputOverlapFinding(ownerProjects(), nil, nil, "")
	if strings.Contains(f.Detail, "models") {
		t.Errorf("models:watch is not a build and is not reported:\n%s", f.Detail)
	}
}

func TestOverlapRemedyShowsTheExactEdits(t *testing.T) {
	f := outputOverlapFinding(ownerProjects(), nil, nil, "[workspace]\nname = \"w\"\n")
	for _, want := range []string{
		`[projects.api.tasks.build-dev] outputs = ["dist-build-dev"]`,
		`change the "build-dev" script in apps/api/package.json`,
		`[workspace] alternative_outputs = [["api:build", "api:build-dev"]]`,
		"--choose 1",
		"--choose 2",
	} {
		if !strings.Contains(f.Remedy, want) {
			t.Errorf("remedy misses %q:\n%s", want, f.Remedy)
		}
	}
}

func TestOverlapChoiceTwoWritesTheAlternativesGroup(t *testing.T) {
	f := outputOverlapFinding(ownerProjects(), nil, [][]string{{"x:a", "x:b"}}, "[workspace]\nname = \"w\"\n")
	if f.Fix == nil || len(f.Fix.Choices) != 1 || len(f.Fix.Choices[0].Options) != 2 {
		t.Fatalf("want one question with two options, got %+v", f.Fix)
	}
	fix, pending, err := doctorfix.Resolve(*f.Fix, []int{2})
	if err != nil || pending != nil {
		t.Fatalf("resolve: %v %v", err, pending)
	}
	got := fix.Files[0].Content
	want := `alternative_outputs = [["x:a", "x:b"], ["api:build", "api:build-dev"]]`
	if !strings.Contains(got, want) {
		t.Errorf("smith.toml misses %q:\n%s", want, got)
	}
	one, _, err := doctorfix.Resolve(*f.Fix, []int{1})
	if err != nil || len(one.Files) != 0 || len(one.Steps) != 1 || one.Steps[0].Argv[len(one.Steps[0].Argv)-1] != "api:build-dev" {
		t.Errorf("choice 1 changes no file and shows the task: %+v %v", one, err)
	}
}

func TestOverlapIsQuietForAlternatives(t *testing.T) {
	f := outputOverlapFinding(ownerProjects(), nil, [][]string{{"api:build", "api:build-dev"}}, "")
	if !f.OK {
		t.Errorf("a pair listed as alternatives is not reported: %+v", f)
	}
}
