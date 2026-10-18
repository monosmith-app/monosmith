package recipestack

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestDeclareAppendsAStackSettingsAndTheRecipeRecord(t *testing.T) {
	root := t.TempDir()
	base := "[workspace]\nname = \"acme\"\n\n# my comment\n[default_stacks]\nnode = \"node-jest\"\n"
	if err := os.WriteFile(filepath.Join(root, "smith.toml"), []byte(base), 0o644); err != nil {
		t.Fatal(err)
	}
	st, _ := StackOf(python())
	st.Name = "team-python-library"
	undo, err := Declare(root, Declaration{
		Project: "billing", Stack: st,
		Settings: map[string]map[string]any{"python": {"layout": "src", "pytest": map[string]any{"cov": true, "paths": []any{"tests", "e2e"}}}},
		Recipe:   workspace.ProjectRecipe{Name: "team", Version: "1.2.0", Source: "recipes/team.recipe.toml", Answers: Answers(python())},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, "smith.toml"))
	if !strings.HasPrefix(string(raw), base) || !strings.Contains(string(raw), "# my comment") {
		t.Fatalf("the existing file was not kept:\n%s", raw)
	}
	cfg, err := workspace.Load(root)
	if err != nil {
		t.Fatalf("%v\n%s", err, raw)
	}
	got := cfg.Stacks["team-python-library"]
	if got.Language != "python" || !reflect.DeepEqual(got.Toolchain, []string{"ruff", "pytest"}) || !reflect.DeepEqual(got.Structure, []string{"vertical-slices"}) {
		t.Errorf("stack %+v", got)
	}
	rec := cfg.Projects["billing"].Recipe
	if rec == nil || rec.Name != "team" || rec.Version != "1.2.0" || rec.Source != "recipes/team.recipe.toml" || rec.Answers["language"] != "python" || rec.Answers["kind"] != "library" {
		t.Errorf("record %+v", rec)
	}
	py := cfg.Projects["billing"].Settings["python"]
	pytest, _ := py["pytest"].(map[string]any)
	if py["layout"] != "src" || pytest["cov"] != true || !reflect.DeepEqual(pytest["paths"], []any{"tests", "e2e"}) {
		t.Errorf("settings %+v", py)
	}

	if err := undo(); err != nil {
		t.Fatal(err)
	}
	if after, _ := os.ReadFile(filepath.Join(root, "smith.toml")); string(after) != base {
		t.Errorf("undo left:\n%s", after)
	}
}

func TestDeclareReusingAStackWritesOnlyTheProject(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "smith.toml"), []byte("[workspace]\nname = \"acme\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	st, _ := StackOf(python())
	st.Name = "team-python-library"
	if _, err := Declare(root, Declaration{Project: "odd name", Stack: st, Reuse: true, Recipe: workspace.ProjectRecipe{Name: "team", Source: "t.toml"}}); err != nil {
		t.Fatal(err)
	}
	cfg, err := workspace.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Stacks) != 0 || cfg.Projects["odd name"].Recipe == nil {
		t.Errorf("stacks %+v, projects %+v", cfg.Stacks, cfg.Projects)
	}
}

func TestDeclareRefusesWhatWouldBreakSmithToml(t *testing.T) {
	root := t.TempDir()
	base := "[workspace]\nname = \"acme\"\n[projects.billing.recipe]\nname = \"x\"\nsource = \"s\"\n"
	if err := os.WriteFile(filepath.Join(root, "smith.toml"), []byte(base), 0o644); err != nil {
		t.Fatal(err)
	}
	st, _ := StackOf(python())
	st.Name = "team-python-library"
	// billing already has a recipe table: appending a second one is a duplicate.
	if _, err := Declare(root, Declaration{Project: "billing", Stack: st, Recipe: workspace.ProjectRecipe{Name: "team", Source: "t.toml"}}); err == nil || !strings.Contains(err.Error(), "would break smith.toml") {
		t.Fatalf("got %v", err)
	}
	if after, _ := os.ReadFile(filepath.Join(root, "smith.toml")); string(after) != base {
		t.Errorf("smith.toml was changed:\n%s", after)
	}
}
