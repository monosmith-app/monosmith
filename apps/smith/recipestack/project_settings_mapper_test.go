package recipestack

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestProjectSettingsKeepsThePluginsTheProjectUses(t *testing.T) {
	res := recipe.Resolution{Settings: map[string]any{
		"python": map[string]any{"layout": "src"},
		"ruff":   map[string]any{"line": 100},
		"dotnet": map[string]any{"target_framework": "net10.0"},
	}}
	st := protocol.Stack{Language: "python", Toolchain: []string{"ruff"}}
	kept, unused := ProjectSettings(res, st)
	if !reflect.DeepEqual(kept, map[string]map[string]any{"python": {"layout": "src"}, "ruff": {"line": 100}}) {
		t.Errorf("kept %v", kept)
	}
	if !reflect.DeepEqual(unused, []string{"dotnet"}) {
		t.Errorf("unused %v", unused)
	}
	if kept, unused := ProjectSettings(recipe.Resolution{}, st); kept != nil || unused != nil {
		t.Errorf("no settings: %v %v", kept, unused)
	}
}
