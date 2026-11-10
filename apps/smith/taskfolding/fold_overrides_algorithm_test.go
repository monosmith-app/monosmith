package taskfolding

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func lib(name string, overrides map[string]workspace.TaskOverride) Project {
	return Project{Name: name, Language: "node", Kind: "lib", Has: map[string]bool{"build": true, "test": true}, Overrides: overrides}
}

func build() map[string]workspace.TaskOverride {
	return map[string]workspace.TaskOverride{"build": {Outputs: []string{"dist"}}}
}

func TestIdenticalLibOverridesBecomeOneGroup(t *testing.T) {
	var projects []Project
	for _, n := range []string{"a", "b", "c", "d"} {
		projects = append(projects, lib(n, build()))
	}
	got := Fold(projects)
	if len(got.Groups) != 1 || strings.Join(got.Groups[0].Select, ",") != "lang:node,kind:lib" {
		t.Fatalf("groups = %+v", got.Groups)
	}
	if !reflect.DeepEqual(got.Groups[0].Tasks["build"], workspace.TaskOverride{Outputs: []string{"dist"}}) {
		t.Errorf("group = %+v", got.Groups[0])
	}
	if len(got.Overrides) != 0 {
		t.Errorf("projects keep %v", got.Overrides)
	}
	if len(got.Folded) != 1 || len(got.Folded[0].Projects) != 4 {
		t.Errorf("folded = %+v", got.Folded)
	}
}

func TestOneLibThatDiffersKeepsItsOwnOverrideWithTheGroupsOtherFields(t *testing.T) {
	same := map[string]workspace.TaskOverride{"test": {DependsOn: []string{"^build"}, Env: map[string]string{"A": "1"}}}
	odd := map[string]workspace.TaskOverride{"test": {Env: map[string]string{"A": "2"}}}
	got := Fold([]Project{lib("a", same), lib("b", same), lib("c", same), lib("odd", odd)})
	if len(got.Groups) != 1 || got.Groups[0].Tasks["test"].Env["A"] != "1" {
		t.Fatalf("groups = %+v", got.Groups)
	}
	want := workspace.TaskOverride{DependsOn: []string{"^build"}, Env: map[string]string{"A": "2"}}
	if o := got.Overrides["odd"]["test"]; !reflect.DeepEqual(o, want) {
		t.Errorf("odd = %+v, want the group's depends_on completed in: %+v", o, want)
	}
	if len(got.Overrides) != 1 {
		t.Errorf("overrides = %v", got.Overrides)
	}
}

func TestAProjectThatWouldTakeValuesItNeverAskedForBlocksTheFold(t *testing.T) {
	got := Fold([]Project{lib("a", build()), lib("b", build()), lib("plain", nil)})
	if len(got.Groups) != 0 {
		t.Errorf("a plain lib with a build task would change: %+v", got.Groups)
	}
}

func TestAProjectWithoutTheTaskDoesNotBlockTheFold(t *testing.T) {
	noBuild := lib("n", nil)
	noBuild.Has = map[string]bool{"test": true}
	got := Fold([]Project{lib("a", build()), lib("b", build()), noBuild})
	if len(got.Groups) != 1 {
		t.Errorf("groups = %+v", got.Groups)
	}
}

func TestASingleProjectIsNotFolded(t *testing.T) {
	if got := Fold([]Project{lib("a", build()), lib("b", nil)}); len(got.Groups) != 0 || len(got.Overrides["a"]) != 1 {
		t.Errorf("result = %+v", got)
	}
}

func TestOtherKindsAreLeftAlone(t *testing.T) {
	app := lib("app", map[string]workspace.TaskOverride{"build": {Outputs: []string{"out"}}})
	app.Kind = "app"
	got := Fold([]Project{lib("a", build()), lib("b", build()), app})
	if len(got.Groups) != 1 || got.Groups[0].Select[len(got.Groups[0].Select)-1] != "kind:lib" {
		t.Fatalf("groups = %+v", got.Groups)
	}
	if len(got.Overrides["app"]) != 1 {
		t.Errorf("the app keeps its override: %v", got.Overrides)
	}
}

// Two projects can share a name (a Node and a Rust project in one folder).
// Folding the first one's last override away must not leave the second one
// writing into a nil map.
func TestProjectsSharingANameDoNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Fold panicked: %v", r)
		}
	}()
	Fold([]Project{lib("a", build()), lib("b", build()), lib("same", build()), lib("same", build())})
}
