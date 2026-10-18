package recipechoice

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

func TestDescribeNamesTheRecipeTheAnswersAndWhatItUses(t *testing.T) {
	a, _ := asker("", false)
	res, err := a.Resolve(load(t), map[string]string{"language": "go"})
	if err != nil {
		t.Fatal(err)
	}
	got := Describe(res, "")
	for _, want := range []string{
		"Recipe team 1.2.0 (language=go, kind=library, cgo=no) will use:",
		"language",
		"go",
		"team.recipe.toml:",
		"setting go.cgo = maybe",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
}

func TestDescribeShowsFilesInTheWorkspaceRelativeToIt(t *testing.T) {
	root := t.TempDir()
	res := recipe.Resolution{Name: "team", Uses: []recipe.Use{
		{Key: "language", Plugins: []string{"go"}, Origin: recipe.Origin{Recipe: "team", File: filepath.Join(root, "recipes", "team.recipe.toml"), Line: 4}},
		{Key: "deploy", Plugins: []string{"docker"}, Origin: recipe.Origin{Recipe: "company", File: "https://example.com/c.toml", Line: 9}},
	}}
	got := Describe(res, root)
	if !strings.Contains(got, "recipes/team.recipe.toml:4 (recipe team)") || !strings.Contains(got, "https://example.com/c.toml:9 (recipe company)") {
		t.Errorf("got\n%s", got)
	}
}
