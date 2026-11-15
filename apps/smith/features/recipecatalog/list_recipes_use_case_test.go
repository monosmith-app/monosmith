package recipecatalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListRecipesFindsRecipeFilesAndSkipsOtherToolsDirectories(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("recipes/team.recipe.toml", "name = \"team\"\nversion = \"1.2.0\"\nextends = \"base.recipe.toml\"\n")
	write("recipes/base.recipe.toml", "name = \"base\"\n")
	write("broken.recipe.toml", "name = [")
	write("node_modules/dep/x.recipe.toml", "name = \"x\"\n")
	write(".git/y.recipe.toml", "name = \"y\"\n")
	write("notes.toml", "name = \"n\"\n")

	got, err := ListRecipes(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].Path != "broken.recipe.toml" || got[0].Error == "" {
		t.Fatalf("listed %+v", got)
	}
	if got[1].Path != "recipes/base.recipe.toml" || got[1].Name != "base" {
		t.Fatalf("listed %+v", got[1])
	}
	if got[2].Name != "team" || got[2].Version != "1.2.0" || got[2].Extends != "base.recipe.toml" {
		t.Fatalf("listed %+v", got[2])
	}
}
