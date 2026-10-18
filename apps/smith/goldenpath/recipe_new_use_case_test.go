//go:build e2e

package goldenpath

import (
	"strings"
	"testing"
)

// answersForTeam is what `smith recipe new` would learn from a person: the
// language is a choice, Node gets a linter and a test runner, Python the
// structure rule that only applies to it.
const answersForTeam = `name = "team"
version = "0.1.0"

[choices.language]
ask = "Language"
options = ["python", "node"]
default = "python"

[levels.language]
choice = "language"
plugins = { python = "python", node = "node" }

[levels.lint]
choice = "language"
plugins = { node = "eslint" }

[levels.test]
choice = "language"
plugins = { node = "vitest" }

[levels.structure]
choice = "language"
plugins = { python = "vertical-slices" }
`

// TestRecipeNew writes a recipe from an answers file with the real binary
// and the real built-in plugins, and checks it for every answer.
func TestRecipeNew(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.write("answers.toml", answersForTeam)

	if out := w.smith(true, "recipe", "new", "--from-answers", "answers.toml", "recipes/team/team.recipe.toml"); !strings.Contains(out, "Wrote recipes/team/team.recipe.toml") {
		t.Fatalf("recipe new:\n%s", out)
	}
	written := w.read("recipes/team/team.recipe.toml")
	for _, want := range []string{`name = "team"`, "[choices.language]", "[when.language.node.use]", `lint = "eslint"`, `test = "vitest"`, "[when.language.python.use]", `structure = "vertical-slices"`} {
		if !strings.Contains(written, want) {
			t.Errorf("the recipe must contain %q:\n%s", want, written)
		}
	}
	if out := w.smith(true, "recipe", "check", "recipes/team/team.recipe.toml", "--all"); !strings.Contains(out, "valid") {
		t.Errorf("recipe check --all:\n%s", out)
	}
	if out := w.smith(false, "recipe", "new", "--from-answers", "answers.toml", "recipes/team/team.recipe.toml"); !strings.Contains(out, "exists") {
		t.Errorf("a second write must be refused:\n%s", out)
	}

	// A recipe that cannot work for one answer is not written.
	w.write("bad.toml", strings.Replace(answersForTeam, `plugins = { python = "vertical-slices" }`, `plugins = { node = "vertical-slices" }`, 1))
	out := w.smith(false, "recipe", "new", "--from-answers", "bad.toml", "bad.recipe.toml")
	if !strings.Contains(out, "nothing was written") || !strings.Contains(out, "language=node") {
		t.Errorf("a recipe that does not resolve:\n%s", out)
	}
}
