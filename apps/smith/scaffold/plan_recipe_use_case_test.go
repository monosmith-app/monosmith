package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func workspaceWithRecipe(t *testing.T, file, body string) *session.Session {
	t.Helper()
	root := t.TempDir()
	if body != "" {
		if err := os.WriteFile(filepath.Join(root, file), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return &session.Session{Root: root, Config: workspace.Config{Workspace: workspace.WorkspaceSection{Name: "acme", Recipe: file}}}
}

func TestPlanRecipeIsNilWhenNoRecipeApplies(t *testing.T) {
	s := &session.Session{Root: t.TempDir()}
	if plan, err := planRecipe(s, AddOptions{Kind: "app", Out: &bytes.Buffer{}}); plan != nil || err != nil {
		t.Fatalf("no recipe: %v, %v", plan, err)
	}
	s = workspaceWithRecipe(t, "team.recipe.toml", "")
	if plan, err := planRecipe(s, AddOptions{Kind: "app", Stack: "node-jest", Out: &bytes.Buffer{}}); plan != nil || err != nil {
		t.Fatalf("--stack must bypass the recipe: %v, %v", plan, err)
	}
}

func TestPlanRecipeRefusesChoicesThatHaveNoRecipeToAnswer(t *testing.T) {
	answers := &RecipeUse{Answers: map[string]string{"language": "python"}}
	s := &session.Session{Root: t.TempDir()}
	if _, err := planRecipe(s, AddOptions{Kind: "app", Recipe: answers, Out: &bytes.Buffer{}}); err == nil || !strings.Contains(err.Error(), "names no recipe") {
		t.Fatalf("no recipe: %v", err)
	}
	s = workspaceWithRecipe(t, "team.recipe.toml", "")
	if _, err := planRecipe(s, AddOptions{Kind: "app", Stack: "node-jest", Recipe: answers, Out: &bytes.Buffer{}}); err == nil || !strings.Contains(err.Error(), "bypasses the recipe") {
		t.Fatalf("--stack and --choice: %v", err)
	}
}

func TestPlanRecipeNamesAMissingRecipeFile(t *testing.T) {
	s := workspaceWithRecipe(t, "recipes/team.recipe.toml", "")
	_, err := planRecipe(s, AddOptions{Kind: "app", Out: &bytes.Buffer{}})
	if err == nil || !strings.Contains(err.Error(), "the recipe recipes/team.recipe.toml") {
		t.Fatalf("got %v", err)
	}
}

func TestPlanRecipeWithoutATerminalNamesTheChoiceItNeeds(t *testing.T) {
	s := workspaceWithRecipe(t, "team.recipe.toml", "name = \"team\"\n[choices.language]\noptions = [\"python\", \"go\"]\n")
	var out bytes.Buffer
	_, err := planRecipe(s, AddOptions{Kind: "package", Recipe: &RecipeUse{}, Out: &out})
	if err == nil || !strings.Contains(err.Error(), `choice "language"`) || !strings.Contains(err.Error(), "--choice language=python|go") {
		t.Fatalf("got %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("a plan that failed must show nothing, showed %q", out.String())
	}
}

func TestPlanRecipeUsesTheRecipeOfTheKind(t *testing.T) {
	s := workspaceWithRecipe(t, "team.recipe.toml", "name = \"team\"\n[choices.language]\noptions = [\"python\"]\n")
	s.Config.Recipes = map[string]string{"lib": "libs.recipe.toml"}
	if err := os.WriteFile(filepath.Join(s.Root, "libs.recipe.toml"), []byte("name = \"libs\"\n[choices.flavour]\noptions = [\"a\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := planRecipe(s, AddOptions{Kind: "lib", Recipe: &RecipeUse{}, Out: &bytes.Buffer{}})
	if err == nil || !strings.Contains(err.Error(), "libs.recipe.toml") || !strings.Contains(err.Error(), `choice "flavour"`) {
		t.Fatalf("the lib recipe was not used: %v", err)
	}
}

func TestSettingsOfProjectKeepsThePluginsTheProjectUses(t *testing.T) {
	res := recipe.Resolution{Settings: map[string]any{
		"python": map[string]any{"layout": "src"},
		"ruff":   map[string]any{"line": 100},
		"dotnet": map[string]any{"target_framework": "net10.0"},
	}}
	st := protocol.Stack{Language: "python", Toolchain: []string{"ruff"}}
	kept, unused := settingsOfProject(res, st)
	if !reflect.DeepEqual(kept, map[string]map[string]any{"python": {"layout": "src"}, "ruff": {"line": 100}}) {
		t.Errorf("kept %v", kept)
	}
	if !reflect.DeepEqual(unused, []string{"dotnet"}) {
		t.Errorf("unused %v", unused)
	}
	if kept, unused := settingsOfProject(recipe.Resolution{}, st); kept != nil || unused != nil {
		t.Errorf("no settings: %v %v", kept, unused)
	}
}

func TestRecipeLocationIsAbsoluteForAPathAndLeavesAURL(t *testing.T) {
	if got := recipeLocation("/ws", "https://example.com/t.toml"); got != "https://example.com/t.toml" {
		t.Errorf("url: %q", got)
	}
	if got := recipeLocation(filepath.FromSlash("/ws"), "recipes/team.toml"); got != filepath.Join(filepath.FromSlash("/ws"), "recipes", "team.toml") {
		t.Errorf("path: %q", got)
	}
}
