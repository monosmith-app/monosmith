package recipewizard

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

// fromAnswers runs `smith recipe new --from-answers` in a folder with no
// workspace around it.
func fromAnswers(t *testing.T, dir, answers, target string, force bool) (string, error) {
	t.Helper()
	t.Chdir(dir)
	file := filepath.Join(dir, "answers.toml")
	if err := os.WriteFile(file, []byte(answers), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := Run(Options{Target: target, FromAnswers: file, Force: force, Out: &out, In: strings.NewReader(""), Lookup: fakeLookup})
	return out.String(), err
}

func TestTheAnswersFileWritesARecipeThatChecksOutForEveryAnswer(t *testing.T) {
	dir := t.TempDir()
	out, err := fromAnswers(t, dir, answersFile, filepath.Join("recipes", "team", "team.recipe.toml"), false)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	written, err := os.ReadFile(filepath.Join(dir, "recipes", "team", "team.recipe.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), "[when.language.node.use]") || !strings.Contains(out, "smith recipe check") {
		t.Errorf("written:\n%s\noutput:\n%s", written, out)
	}
	doc, err := recipe.ParseRecipe(written, recipe.Source{Ref: "team.recipe.toml"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recipe.ValidateAll(recipe.Loaded{Chain: []recipe.Document{doc}}, fakeLookup); err != nil {
		t.Errorf("the recipe the wizard wrote does not check out: %v", err)
	}
}

func TestNothingIsWrittenForARecipeThatDoesNotResolve(t *testing.T) {
	// ruff needs python, and the node branch chooses it.
	bad := strings.Replace(answersFile, `plugins = { python = "ruff", node = "eslint" }`, `plugins = { python = "ruff", node = "ruff" }`, 1)
	dir := t.TempDir()
	_, err := fromAnswers(t, dir, bad, "team.recipe.toml", false)
	if err == nil || !strings.Contains(err.Error(), "nothing was written") || !strings.Contains(err.Error(), "language=node") {
		t.Fatalf("err = %v; it should say nothing was written and for which answer", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "team.recipe.toml")); statErr == nil {
		t.Error("a recipe that does not resolve was written")
	}
}

func TestAnExistingRecipeIsReplacedOnlyOnRequest(t *testing.T) {
	dir := t.TempDir()
	if _, err := fromAnswers(t, dir, answersFile, "team.recipe.toml", false); err != nil {
		t.Fatal(err)
	}
	if _, err := fromAnswers(t, dir, answersFile, "team.recipe.toml", false); !errors.Is(err, ErrExists) {
		t.Errorf("err = %v; the second write must be refused", err)
	}
	if _, err := fromAnswers(t, dir, answersFile, "team.recipe.toml", true); err != nil {
		t.Errorf("--force: %v", err)
	}
}

func TestAnExistingRecipeIsReplacedWhenThePersonConfirms(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := fromAnswers(t, dir, answersFile, "team.recipe.toml", false); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	script := strings.Join([]string{"team", "", "", "c", "", "", "python, node", "python", "2", "1", "", "y"}, "\n") + "\n"
	err := Run(Options{Target: "team.recipe.toml", Interactive: true, Out: &out, In: strings.NewReader(script), Lookup: fakeLookup, Names: []string{"node", "python"}})
	if err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Replace it?") || !strings.Contains(out.String(), "Wrote team.recipe.toml") {
		t.Errorf("output:\n%s", out.String())
	}
}

func TestTheWizardNeedsATerminalOrAnAnswersFile(t *testing.T) {
	t.Chdir(t.TempDir())
	err := Run(Options{Out: &bytes.Buffer{}, In: strings.NewReader(""), Lookup: fakeLookup})
	if err == nil || !strings.Contains(err.Error(), "--from-answers") {
		t.Errorf("err = %v", err)
	}
}

func TestAParentIsLoadedAndItsChoicesAreUsed(t *testing.T) {
	dir := t.TempDir()
	base := `name = "base"
version = "1.0.0"

[choices.language]
options = ["python", "node"]
default = "python"

[when.language.python.use]
language = "python"

[when.language.node.use]
language = "node"
`
	if err := os.WriteFile(filepath.Join(dir, "base.recipe.toml"), []byte(base), 0o644); err != nil {
		t.Fatal(err)
	}
	child := `
name = "team"
version = "0.2.0"
extends = "base.recipe.toml"

[levels.lint]
choice = "language"
plugins = { python = "ruff", node = "eslint" }
`
	out, err := fromAnswers(t, dir, child, "team.recipe.toml", false)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	written, _ := os.ReadFile(filepath.Join(dir, "team.recipe.toml"))
	if !strings.Contains(string(written), `extends = "base.recipe.toml"`) || strings.Contains(string(written), "[choices.") {
		t.Errorf("written:\n%s", written)
	}
}
