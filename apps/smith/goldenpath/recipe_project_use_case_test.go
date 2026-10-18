//go:build e2e

package goldenpath

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestRecipeProject releases a recipe like a package: the recipe is a project
// of its own, versioned from the commits that touch it, with its version
// written into the file and tagged name@x.y.z.
func TestRecipeProject(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-recipes", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-recipes")
	remote := filepath.Join(base, "remote.git")
	run(t, base, "git", "init", "-q", "--bare", remote)
	w.git("remote", "add", "origin", remote)

	// The recipe is added like any project, and is one.
	w.smith(true, "add", "package", "team", "--lang", "recipe")
	if out := w.smith(true, "projects"); !strings.Contains(out, "team") || !strings.Contains(out, "recipe") {
		t.Fatalf("projects:\n%s", out)
	}
	if recipe := w.read("packages/team/team.recipe.toml"); !strings.Contains(recipe, `name = "team"`) || !strings.Contains(recipe, `version = "0.0.0"`) {
		t.Fatalf("the new recipe:\n%s", recipe)
	}
	// Its test task checks every answer, through the recipe's own command.
	if out := w.smith(true, "run", "test", "-p", "team"); !strings.Contains(out, "team:test") {
		t.Fatalf("test:\n%s", out)
	}
	w.write("packages/team/team.recipe.toml", strings.Replace(w.read("packages/team/team.recipe.toml"), `language = "python"`, `language = "cobol"`, 1))
	if out := w.smith(false, "run", "test", "-p", "team"); !strings.Contains(out, "cobol") {
		t.Fatalf("a recipe naming a plugin that does not exist must fail its test:\n%s", out)
	}
	w.write("packages/team/team.recipe.toml", strings.Replace(w.read("packages/team/team.recipe.toml"), `language = "cobol"`, `language = "python"`, 1))

	// The first release writes the version into the file, tags it, and
	// attaches the file.
	w.git("add", "-A")
	w.git("commit", "-qm", "feat(team): the first recipe")
	if out := w.smith(true, "release", "--dry-run"); !strings.Contains(out, "team") || !strings.Contains(out, "0.1.0") || !strings.Contains(out, "dist/release/team/team.recipe.toml") {
		t.Fatalf("the preview names the version and the file it attaches:\n%s", out)
	}
	if out := w.smith(true, "release", "--yes"); !strings.Contains(out, "tagged team@0.1.0") {
		t.Fatalf("release:\n%s", out)
	}
	// A release commits nothing: the recipe in the repository keeps its
	// version, and the copy attached to the release says which release it is.
	if recipe := w.read("packages/team/team.recipe.toml"); !strings.Contains(recipe, `version = "0.0.0"`) {
		t.Fatalf("the recipe in the repository changed:\n%s", recipe)
	}
	if asset := w.read("dist/release/team/team.recipe.toml"); !strings.Contains(asset, `version = "0.1.0"`) || !strings.Contains(asset, "[choices.language]") {
		t.Fatalf("the attached recipe carries the release's version and the rest of the file:\n%s", asset)
	}
	if tags := w.git("ls-remote", "--tags", "origin"); !strings.Contains(tags, "team@0.1.0") {
		t.Fatalf("the tag is pushed:\n%s", tags)
	}

	// A fix releases a patch.
	w.write("packages/team/team.recipe.toml", strings.Replace(w.read("packages/team/team.recipe.toml"), `default = "node"`, `default = "python"`, 1))
	w.git("add", "-A")
	w.git("commit", "-qm", "fix(team): python is the default")
	if out := w.smith(true, "release", "--dry-run"); !strings.Contains(out, "0.1.1") {
		t.Fatalf("a fix is a patch:\n%s", out)
	}
}
