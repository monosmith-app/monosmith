package upgrade

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func recipeServer(t *testing.T, body *string) *httptest.Server {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(*body))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestRecipeChangesAcceptsAChangedURLRecipe(t *testing.T) {
	root := t.TempDir()
	body := "name = \"team\"\n[use]\nlanguage = \"go\"\n"
	server := recipeServer(t, &body)
	loader := &recipe.Loader{Root: root, Client: server.Client()}
	url := server.URL + "/team.recipe.toml"
	cfg := workspace.Config{Workspace: workspace.WorkspaceSection{Recipe: url}}

	if _, err := loader.Load(url); err != nil {
		t.Fatal(err)
	}
	if got, err := recipeChanges(loader, cfg, false); err != nil || len(got) != 0 {
		t.Fatalf("unchanged recipe: %v, %v", got, err)
	}

	// A pinned recipe is read from the cache without the network; a fresh
	// clone or CI has no cache and reads the URL, which must match the pin.
	body = "name = \"team\"\n[use]\nlanguage = \"python\"\n"
	if err := os.RemoveAll(filepath.Join(root, ".smith", "recipes")); err != nil {
		t.Fatal(err)
	}
	if _, err := loader.Load(url); err == nil {
		t.Fatal("a changed recipe was loaded without being accepted")
	}
	got, err := recipeChanges(loader, cfg, true)
	if err != nil || len(got) != 1 || !strings.Contains(got[0], "changed since it was pinned") {
		t.Fatalf("check: %v, %v", got, err)
	}
	if _, err := loader.Load(url); err == nil {
		t.Fatal("check accepted the change")
	}
	got, err = recipeChanges(loader, cfg, false)
	if err != nil || len(got) != 1 || !strings.Contains(got[0], "accepted") {
		t.Fatalf("accept: %v, %v", got, err)
	}
	if _, err := loader.Load(url); err != nil {
		t.Fatalf("after accepting: %v", err)
	}
}

func TestRecipeChangesLeavesPathsAndNoRecipe(t *testing.T) {
	loader := &recipe.Loader{Root: t.TempDir()}
	if got, err := recipeChanges(loader, workspace.Config{}, false); err != nil || got != nil {
		t.Errorf("no recipe: %v, %v", got, err)
	}
	cfg := workspace.Config{Workspace: workspace.WorkspaceSection{Recipe: "recipes/team.recipe.toml"}}
	if got, err := recipeChanges(loader, cfg, false); err != nil || len(got) != 0 {
		t.Errorf("path recipe: %v, %v", got, err)
	}
}
