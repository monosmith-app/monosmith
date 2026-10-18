package recipestack

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func python() recipe.Resolution {
	return recipe.Resolution{
		Name: "team", Version: "1.2.0",
		Asked: []recipe.Question{{Name: "language", Answer: "python"}, {Name: "kind", Answer: "library"}},
		Uses: []recipe.Use{
			{Key: "language", Plugins: []string{"python"}},
			{Key: "lint", Plugins: []string{"ruff@^1"}},
			{Key: "structure", Plugins: []string{"vertical-slices"}},
			{Key: "test", Plugins: []string{"pytest"}},
		},
	}
}

func TestStackOfMapsLevelsAndSlots(t *testing.T) {
	st, err := StackOf(python())
	if err != nil {
		t.Fatal(err)
	}
	want := protocol.Stack{
		Description: "From the recipe team (language=python, kind=library); written by smith",
		Language:    "python", Toolchain: []string{"ruff", "pytest"}, Structure: []string{"vertical-slices"},
	}
	if !reflect.DeepEqual(st, want) {
		t.Fatalf("got %+v want %+v", st, want)
	}
}

func TestStackOfKeepsAnExplicitToolchainOrderAndDeduplicates(t *testing.T) {
	res := recipe.Resolution{Name: "t", Uses: []recipe.Use{
		{Key: "language", Plugins: []string{"node"}},
		{Key: "scaffold", Plugins: []string{"typescript"}},
		{Key: "toolchain", Plugins: []string{"typescript", "eslint", "jest"}},
		{Key: "lint", Plugins: []string{"eslint"}},
		{Key: "deploy", Plugins: []string{"docker"}},
	}}
	st, err := StackOf(res)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(st.Toolchain, []string{"typescript", "eslint", "jest"}) || st.Scaffold != "typescript" || !reflect.DeepEqual(st.Deploy, []string{"docker"}) {
		t.Fatalf("got %+v", st)
	}
}

func TestStackOfNeedsALanguage(t *testing.T) {
	if _, err := StackOf(recipe.Resolution{Name: "t", Uses: []recipe.Use{{Key: "lint", Plugins: []string{"ruff"}}}}); err == nil || !strings.Contains(err.Error(), "chooses no language") {
		t.Fatalf("got %v", err)
	}
}
