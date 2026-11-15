package guide

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewRecipeAnswersBecomeFlags(t *testing.T) {
	got := NewRecipeArgv(map[string][]string{
		"name": {"team"}, "version": {"1.2.0"}, "extends": {"base.recipe.toml"}, "fixed": {"language=python  lint=ruff"},
	})
	want := []string{"recipe", "new", "--name", "team", "--recipe-version", "1.2.0", "--extends", "base.recipe.toml", "--fixed", "language=python", "--fixed", "lint=ruff"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestNewRecipeShortestCommandHasOnlyTheName(t *testing.T) {
	got := NewRecipeArgv(map[string][]string{"name": {"team"}, "version": {""}, "extends": {""}, "fixed": {""}})
	if want := []string{"recipe", "new", "--name", "team"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestNewRecipeStepsRefuseWhatTheCommandRefuses(t *testing.T) {
	check := map[string]func(map[string][]string, string) string{}
	for _, s := range NewRecipeSteps() {
		if s.Check != nil {
			check[s.Key] = s.Check
		}
	}
	if check["name"](nil, "Bad Name") == "" || check["name"](nil, "team") != "" {
		t.Error("the name check disagrees with the command")
	}
	if check["version"](nil, "1.0") == "" || check["version"](map[string][]string{"name": {"team"}}, "") != "" {
		t.Error("the version check disagrees with the command")
	}
	a := map[string][]string{"name": {"team"}}
	if got := check["fixed"](a, "python"); !strings.Contains(got, "role=plugin") {
		t.Errorf("a fixed value without a level was accepted: %q", got)
	}
	if got := check["fixed"](a, "language=python,node"); got == "" {
		t.Error("two plugins for the language level were accepted")
	}
	if got := check["fixed"](a, "language=python structure=slices,layers"); got != "" {
		t.Errorf("a good fixed value was refused: %q", got)
	}
}
