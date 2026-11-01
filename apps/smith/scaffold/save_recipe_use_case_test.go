package scaffold

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/architecture"
	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestARecipeSavedFromTheAnswersReadsBackAsTheSameAnswers(t *testing.T) {
	for _, layout := range architecture.Names() {
		opts := NewOptions{Version: "1.0.0", Language: "node", Architecture: layout, DefaultStacks: map[string]string{"node": "node-vitest"}}
		text, err := renderRecipe(opts, "team")
		if err != nil {
			t.Fatalf("%s: %v", layout, err)
		}
		file := filepath.Join(t.TempDir(), "team.recipe.toml")
		if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		loaded, err := (&recipe.Loader{}).Load(file)
		if err != nil {
			t.Fatalf("%s: %v", layout, err)
		}
		want := map[string]string{AnswerLanguage: "node", AnswerNodeStack: "node-vitest"}
		if layout != "none" {
			want[AnswerArchitecture] = layout
		}
		if got := answersFromRecipe(loaded); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: recipe answers %v, want %v\n%s", layout, got, want, text)
		}
	}
}

func TestEveryOfferedLanguageMakesAValidRecipe(t *testing.T) {
	for _, q := range NewQuestions() {
		if q.Key != AnswerLanguage {
			continue
		}
		for _, o := range q.Options {
			if o.Value == LanguageLater {
				continue
			}
			if _, err := renderRecipe(NewOptions{Version: "1.0.0", Language: o.Value}, "team"); err != nil {
				t.Errorf("%s: %v", o.Value, err)
			}
		}
	}
}

func TestNoLanguageNoRecipe(t *testing.T) {
	if _, err := renderRecipe(NewOptions{Version: "1.0.0"}, "team"); err == nil || !strings.Contains(err.Error(), "main language") {
		t.Fatalf("%v", err)
	}
}

func TestSavingARecipeAtTheEndWritesItInTheFirstCommit(t *testing.T) {
	dir, out, err := runNew(t, true, "\nnode\n2\n\n\n\nyes\nteam\n")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Save these answers as a recipe to reuse on other workspaces? [y/N]:", "Name of the recipe [acme]:", "wrote recipes/team.recipe.toml", "For another workspace: smith new <name> --recipe-file"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
	if strings.Index(out, "Your choices") > strings.Index(out, "Save these answers") {
		t.Errorf("the question comes after the summary:\n%s", out)
	}
	if _, err := (&recipe.Loader{}).Load(filepath.Join(dir, "recipes", "team.recipe.toml")); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "committed") {
		t.Errorf("not committed:\n%s", out)
	}
}

func TestNotSavingIsTheDefaultAndNeedsALanguage(t *testing.T) {
	dir, _, err := runNew(t, true, "\nnode\n\n\n\n\n")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "recipes")); err == nil {
		t.Error("a recipe was saved without a yes")
	}
	_, out, err := runNew(t, true, "\n\n\n\n")
	if err != nil || strings.Contains(out, "No recipe to save") || strings.Contains(out, "Save these answers") {
		t.Errorf("%v\n%s", err, out)
	}
}

func TestANewWorkspaceFromASavedRecipeSkipsWhatItAnswers(t *testing.T) {
	first, _, err := runNew(t, false, "", "--language", "node", "--default-stack", "node=node-vitest", "--architecture", "hexagonal", "--save-recipe", "team")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(first, "recipes", "team.recipe.toml")

	dir, out, err := runNew(t, true, "\n\n", "--recipe-file", file)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, asked := range []string{"main language of this workspace", "Which tests and lint", "Which code layout"} {
		if strings.Contains(out, asked) {
			t.Errorf("asked %q although the recipe answers it:\n%s", asked, out)
		}
	}
	if !strings.Contains(out, "The recipe already answers the main language, the Node tests and lint, the code layout.") || !strings.Contains(out, "Which CI") {
		t.Errorf("\n%s", out)
	}
	cfg, err := workspace.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Workspace.Recipe != "recipes/team.recipe.toml" || cfg.DefaultStacks["node"] != "node-vitest" {
		t.Errorf("%+v %v", cfg.Workspace, cfg.DefaultStacks)
	}
	if _, ok := cfg.Conventions["hexagonal"]; !ok {
		t.Errorf("conventions %v", cfg.Conventions)
	}
	if _, err := os.Stat(filepath.Join(dir, "recipes", "team.recipe.toml")); err != nil {
		t.Error("the recipe was not copied in")
	}
}

func TestARecipeThatIsAlreadyThereIsNotReplaced(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "recipes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "recipes", "team.recipe.toml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	opts := NewOptions{Version: "1.0.0", Language: "go", SaveRecipe: "team"}
	if _, err := recipeFiles(dir, &opts); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("%v", err)
	}
}
