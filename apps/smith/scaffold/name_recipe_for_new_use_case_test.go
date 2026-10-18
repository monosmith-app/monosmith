package scaffold

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestWithRecipeWritesTheSettingAfterTheModulePrefix(t *testing.T) {
	config := fmt.Sprintf(configTemplate, "acme", "@acme", "github.com/acme/acme", `"github"`)
	got := withRecipe(config, "github.com/acme/acme", "recipes/team.recipe.toml")
	cfg, err := workspace.Parse([]byte(got))
	if err != nil {
		t.Fatalf("%v\n%s", err, got)
	}
	if cfg.Workspace.Recipe != "recipes/team.recipe.toml" || cfg.Workspace.ModulePrefix != "github.com/acme/acme" {
		t.Errorf("workspace %+v", cfg.Workspace)
	}
	if strings.Index(got, "recipe =") < strings.Index(got, "module_prefix") || strings.Index(got, "recipe =") > strings.Index(got, "[ci]") {
		t.Errorf("the setting is not in [workspace]:\n%s", got)
	}
}

func TestCheckRecipe(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	if err := checkRecipe(dir, "", &out); err != nil {
		t.Errorf("none: %v", err)
	}
	if err := checkRecipe(dir, "../shared/team.toml", &out); err == nil || !strings.Contains(err.Error(), "--recipe") {
		t.Errorf("a path out of the workspace: %v", err)
	}
	if err := checkRecipe(dir, "http://example.com/t.toml", &out); err == nil {
		t.Error("plain http was accepted")
	}

	// A recipe that is not there yet is said, not refused.
	out.Reset()
	if err := checkRecipe(dir, "recipes/team.recipe.toml", &out); err != nil || !strings.Contains(out.String(), "not there yet") {
		t.Errorf("missing file: %v, %q", err, out.String())
	}

	// One that is there must be a recipe.
	if err := os.MkdirAll(filepath.Join(dir, "recipes"), 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "recipes", "team.recipe.toml")
	if err := os.WriteFile(file, []byte("name = \"team\"\nnope = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := checkRecipe(dir, "recipes/team.recipe.toml", &out); err == nil || !strings.Contains(err.Error(), "--recipe recipes/team.recipe.toml") {
		t.Errorf("a file that is not a recipe: %v", err)
	}
	if err := os.WriteFile(file, []byte("name = \"team\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := checkRecipe(dir, "recipes/team.recipe.toml", &out); err != nil {
		t.Errorf("a recipe: %v", err)
	}
}
