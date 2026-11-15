package engineserver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func recipeWorkspace(t *testing.T) *client {
	t.Helper()
	t.Setenv("SMITH_WORKSPACE_ROOT", "") // smith sets it for the tasks it runs
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "smith.toml"), []byte("schema = 1\n\n[workspace]\nname = \"acme\"\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(root, "recipes"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "recipes", "team.recipe.toml"), []byte(teamRecipeFile), 0o644)
	return start(t, &Server{Version: "1", Dir: root})
}

func TestRecipeListFindsTheWorkspacesRecipes(t *testing.T) {
	c := recipeWorkspace(t)
	recipes := c.call("recipe.list", nil)["result"].(map[string]any)["recipes"].([]any)
	if len(recipes) != 1 {
		t.Fatalf("recipe.list answered %v", recipes)
	}
	if r := recipes[0].(map[string]any); r["path"] != "recipes/team.recipe.toml" || r["name"] != "team" || r["version"] != "1.2.0" {
		t.Fatalf("recipe.list answered %v", r)
	}
}

func TestRecipeDescribeShowsWhatARecipeOffersBeforeAnyAnswer(t *testing.T) {
	c := recipeWorkspace(t)
	d := c.call("recipe.describe", map[string]any{"recipe": "recipes/team.recipe.toml"})["result"].(map[string]any)
	if d["name"] != "team" || len(d["choices"].([]any)) != 1 || fmt.Sprint(d["configures"]) != "[dotnet]" {
		t.Fatalf("recipe.describe answered %v", d)
	}
	var language map[string]any
	for _, u := range d["use"].([]any) {
		if u.(map[string]any)["key"] == "language" {
			language = u.(map[string]any)
		}
	}
	if fmt.Sprint(language["alternatives"]) != "map[language=csharp:[dotnet] language=python:[python]]" {
		t.Fatalf("language role %v", language)
	}
}

func TestRecipeRolesSaysWhatTheRecipePutsInEachRoleForTheAnswers(t *testing.T) {
	c := recipeWorkspace(t)
	r := c.call("recipe.roles", map[string]any{"recipe": "recipes/team.recipe.toml", "choices": map[string]string{"language": "csharp"}})["result"].(map[string]any)
	got := map[string]string{}
	for _, role := range r["roles"].([]any) {
		m := role.(map[string]any)
		got[m["key"].(string)] = fmt.Sprint(m["plugins"])
	}
	if fmt.Sprint(got) != "map[language:[dotnet] structure:[vertical-slices]]" {
		t.Fatalf("roles %v", got)
	}
	if fmt.Sprint(r["choices"].([]any)[0].(map[string]any)["answer"]) != "csharp" {
		t.Fatalf("choices %v", r["choices"])
	}
	// A choice without a default has to be answered.
	missing := c.call("recipe.roles", map[string]any{"recipe": "recipes/team.recipe.toml"})
	if msg := missing["error"].(map[string]any)["message"].(string); !strings.Contains(msg, "language") {
		t.Fatalf("an unanswered choice answered %v", missing)
	}
}

func TestARecipeOutsideTheWorkspaceIsRefused(t *testing.T) {
	c := recipeWorkspace(t)
	for _, ref := range []string{"../team.recipe.toml", "--all", "http://example.com/x.recipe.toml"} {
		reply := c.call("recipe.describe", map[string]any{"recipe": ref})
		if reply["error"] == nil {
			t.Errorf("%s answered %v", ref, reply)
		}
	}
}
