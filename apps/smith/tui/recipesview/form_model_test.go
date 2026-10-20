package recipesview

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

const teamRecipe = `name = "team"
version = "1.2.0"

[choices.language]
ask = "Language"
options = ["python", "csharp"]

[choices.kind]
ask = "What are you creating?"
options = ["library", "service"]
default = "library"

[use]
structure = "vertical-slices"

[when.language.python.use]
language = "python"

[when.language.csharp.use]
language = "dotnet"

[when.language.csharp.settings.dotnet]
target_framework = "net10.0"

[when.language.csharp.when.kind.service.use]
deploy = "docker"
`

// workspaceWithRecipes writes a workspace whose apps are made from team.recipe.toml.
func workspaceWithRecipes(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"smith.toml":                   "schema = 1\n\n[workspace]\nname = \"acme\"\n\n[recipes]\napp = \"recipes/team.recipe.toml\"\n",
		"recipes/team.recipe.toml":     teamRecipe,
		"recipes/broken.recipe.toml":   "this is not = = toml",
		"recipes/lonely.recipe.toml":   "name = \"lonely\"\nversion = \"0.1.0\"\n\n[use]\nlanguage = \"go\"\n",
		"recipes/ignored.txt":          "not a recipe",
		"node_modules/x/y.recipe.toml": "name = \"skipped\"\n",
	}
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// form is a model on a workspace, with every description it asks for loaded.
type form struct {
	t    *testing.T
	root string
	m    Model
}

func newForm(t *testing.T) *form {
	root := workspaceWithRecipes(t)
	f := &form{t: t, root: root, m: New().FromLoaded(Loaded{Catalog: ReadCatalog(root)})}
	return f.settle()
}

// settle answers every request the model makes, as the shell would.
func (f *form) settle() *form {
	for range 5 {
		key, ref, answers, ok := f.m.Request()
		if !ok {
			return f
		}
		f.m = f.m.Requested(key)
		f.m = f.m.Receive(DetailLoaded{Key: key, Detail: Describe(f.root, ref, answers)})
	}
	f.t.Fatal("the model keeps asking for descriptions")
	return f
}

func (f *form) keys(names ...string) *form {
	for _, n := range names {
		var k tea.KeyMsg
		switch n {
		case "up":
			k = tea.KeyMsg{Type: tea.KeyUp}
		case "down":
			k = tea.KeyMsg{Type: tea.KeyDown}
		case "left":
			k = tea.KeyMsg{Type: tea.KeyLeft}
		case "right":
			k = tea.KeyMsg{Type: tea.KeyRight}
		case "enter":
			k = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			k = tea.KeyMsg{Type: tea.KeyEsc}
		default:
			k = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(n)}
		}
		f.m = f.m.Update(k)
		f.settle()
	}
	return f
}

func (f *form) typeText(s string) *form {
	f.m = f.m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)})
	return f
}

func (f *form) screen() string {
	lines, _ := f.m.View(theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII}), 140)
	return strings.Join(lines, "\n")
}

func (f *form) has(texts ...string) {
	f.t.Helper()
	out := f.screen()
	for _, want := range texts {
		if !strings.Contains(out, want) {
			f.t.Fatalf("missing %q on screen:\n%s", want, out)
		}
	}
}

func (f *form) hasNot(texts ...string) {
	f.t.Helper()
	out := f.screen()
	for _, text := range texts {
		if strings.Contains(out, text) {
			f.t.Fatalf("unexpected %q on screen:\n%s", text, out)
		}
	}
}

func (f *form) argv() []string {
	argv, _ := f.m.Argv()
	return argv
}

func TestTheCatalogListsRecipesAndWhichRecipeMakesEachKind(t *testing.T) {
	c := ReadCatalog(workspaceWithRecipes(t))
	if c.Err != nil {
		t.Fatal(c.Err)
	}
	var paths []string
	for _, l := range c.Recipes {
		paths = append(paths, l.Path)
	}
	want := []string{"recipes/broken.recipe.toml", "recipes/lonely.recipe.toml", "recipes/team.recipe.toml"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("recipes %v, want %v", paths, want)
	}
	if c.ByKind["app"] != "recipes/team.recipe.toml" || c.ByKind["lib"] != "" {
		t.Fatalf("by kind %v", c.ByKind)
	}
}

func TestANewProjectShowsWhatTheRecipePutsInEachRoleBeforeAnythingIsWritten(t *testing.T) {
	f := newForm(t)
	f.has("Create", "new project", "app", "made from recipes/team.recipe.toml", "Recipe team 1.2.0")
	// Language has no default, so the roles wait for an answer.
	f.has("Roles depend on language", "needs an answer", "python if language=python", "dotnet if language=csharp")
	if argv, reason := f.m.Argv(); argv != nil || !strings.Contains(reason, "type the project's name") {
		t.Fatalf("argv %v reason %q", argv, reason)
	}
	f.keys("down", "down", "enter").typeText("web").keys("enter")
	if argv, reason := f.m.Argv(); argv != nil || !strings.Contains(reason, "answer language first") {
		t.Fatalf("an unanswered choice must stop the command: %v %q", argv, reason)
	}
	f.keys("down", "right") // language: the first option, python
	f.has("It puts, for these answers:", "language", "python", "structure", "vertical-slices", "recipes/team.recipe.toml:", "(recipe team)")
	f.hasNot("Roles depend on")
	if got, want := f.argv(), []string{"add", "app", "web", "--choice", "language=python"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("argv %v, want %v", got, want)
	}
	f.has("smith add app web --choice language=python", "c creates it after you confirm")
}

func TestAnAnswerThatOpensAChoiceAddsItsRowAndChangesTheRoles(t *testing.T) {
	f := newForm(t).keys("down", "down", "enter").typeText("api").keys("enter", "down", "right", "right") // language: csharp
	f.has("dotnet", "settings for: dotnet", "kind")
	f.has("(default)") // kind keeps its default until it is changed
	f.hasNot("docker")
	f.keys("down", "right") // kind: service
	f.has("docker")
	want := []string{"add", "app", "api", "--choice", "language=csharp", "--choice", "kind=service"}
	if got := f.argv(); !reflect.DeepEqual(got, want) {
		t.Fatalf("argv %v, want %v", got, want)
	}
}

func TestAnAnswerTheNewAnswersMakeIrrelevantIsLeftOut(t *testing.T) {
	f := newForm(t).keys("down", "down", "enter").typeText("api").keys("enter", "down", "right", "right", "down", "right") // csharp, kind=service
	f.keys("up", "left")                                                                                                   // language back to python: kind is still asked at the top of the recipe
	got := f.argv()
	want := []string{"add", "app", "api", "--choice", "language=python", "--choice", "kind=service"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv %v, want %v", got, want)
	}
	f.has("python")
	f.hasNot("docker") // the service block only applies to csharp
}

func TestAKindWithoutARecipeSaysHowToGetOne(t *testing.T) {
	f := newForm(t).keys("down", "right") // lib
	f.has("no recipe: made by the stack or the language plugin", "No recipe")
	if argv, reason := f.m.Argv(); argv != nil || !strings.Contains(reason, "no recipe makes libs") {
		t.Fatalf("argv %v reason %q", argv, reason)
	}
}

func TestANewWorkspaceStartsFromARecipeOfTheWorkspaceOrNone(t *testing.T) {
	f := newForm(t).keys("right") // mode: new workspace
	f.has("new workspace", "Recipe", "none")
	if argv, reason := f.m.Argv(); argv != nil || !strings.Contains(reason, "workspace's name") {
		t.Fatalf("argv %v reason %q", argv, reason)
	}
	f.keys("down", "enter").typeText("shop").keys("enter")
	if got, want := f.argv(), []string{"new", "shop"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("argv %v, want %v", got, want)
	}
	f.keys("down", "right", "right") // none -> lonely (broken is skipped) -> team
	if got, want := f.argv(), []string{"new", "shop", "--recipe", "recipes/team.recipe.toml"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("argv %v, want %v", got, want)
	}
	f.has("Recipe team 1.2.0", "Roles depend on language", "Choices are asked when a project is added")
	f.keys("left") // lonely: nothing to ask
	f.has("Recipe lonely 0.1.0", "language", "go")
}

func TestARecipeWithNoChoicesShowsItsRolesAtOnce(t *testing.T) {
	f := newForm(t).keys("right", "down", "down", "right") // workspace, recipe row: lonely
	f.has("language", "go", "recipes/lonely.recipe.toml:5")
}

func TestTheRecipeWizardHandsOverToSmithRecipeNewInATerminal(t *testing.T) {
	f := newForm(t).keys("left") // mode wraps to new recipe
	f.has("new recipe", "The recipe wizard", "smith recipe new asks", "team.recipe.toml")
	if got, want := f.argv(), []string{"recipe", "new", DefaultRecipePath}; !reflect.DeepEqual(got, want) {
		t.Fatalf("argv %v, want %v", got, want)
	}
	f.has("smith recipe new team.recipe.toml", "in a terminal of its own")
	f.keys("down", "enter")
	for range len(DefaultRecipePath) {
		f.m = f.m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	f.typeText("recipes/acme.recipe.toml").keys("enter")
	if got, want := f.argv(), []string{"recipe", "new", "recipes/acme.recipe.toml"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("argv %v, want %v", got, want)
	}
	f.keys("down", "enter") // path again, then clear it
	for range 30 {
		f.m = f.m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	f.keys("enter")
	if argv, reason := f.m.Argv(); argv != nil || !strings.Contains(reason, "where to write") {
		t.Fatalf("an empty path must stop the command: %v %q", argv, reason)
	}
}

func TestEscapeLeavesAFieldAsItWas(t *testing.T) {
	f := newForm(t).keys("down", "down", "enter").typeText("oops").keys("esc")
	if f.m.Name != "" || f.m.Editing() {
		t.Fatalf("name %q editing %v", f.m.Name, f.m.Editing())
	}
}

func TestABrokenRecipeIsReportedNotResolved(t *testing.T) {
	root := workspaceWithRecipes(t)
	d := Describe(root, "recipes/broken.recipe.toml", nil)
	if d.Problem == "" || len(d.Roles) != 0 {
		t.Fatalf("%+v", d)
	}
	if d := Describe(root, "recipes/missing.recipe.toml", nil); d.Problem == "" {
		t.Fatal("a missing recipe is a problem")
	}
	if d := Describe(root, "", nil); d.Problem != "" || d.Ref != "" {
		t.Fatalf("no recipe is not a problem: %+v", d)
	}
}

func TestADescriptionThatArrivesAfterTheFormMovedOnIsDropped(t *testing.T) {
	f := newForm(t)
	key, ref, answers, ok := f.m.Request()
	_ = ref
	_ = answers
	if ok {
		t.Fatalf("nothing is wanted once settled: %q", key)
	}
	f.keys("down", "down", "down", "right") // answer language
	stale := DetailLoaded{Key: "recipes/team.recipe.toml\x00language=csharp", Detail: Detail{Ref: "recipes/team.recipe.toml", Name: "stale"}}
	f.m = f.m.Receive(stale)
	f.hasNot("stale")
}

func TestTheFormFitsNarrowScreensAndNamesTheUnreadableRecipes(t *testing.T) {
	f := newForm(t)
	lines, rowOf := f.m.View(theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII}), 60)
	for i, l := range lines {
		if len(l) > 60*4 { // generous: plain text, ASCII only, no styling
			t.Fatalf("line %d is too long: %q", i, l)
		}
	}
	if len(rowOf) != len(f.m.rows()) {
		t.Fatalf("every row is drawn once: %d lines for %d rows", len(rowOf), len(f.m.rows()))
	}
	f.has("Recipes in this workspace (3)", "recipes/broken.recipe.toml")
}

func TestClickingARowSelectsItAndClickingTheSelectedRowActs(t *testing.T) {
	f := newForm(t)
	f.m = f.m.SelectRow(1)
	if f.m.Cursor != 1 || f.m.Kind != "app" {
		t.Fatalf("cursor %d kind %s", f.m.Cursor, f.m.Kind)
	}
	f.m = f.m.SelectRow(1) // acts: the kind moves on
	if f.m.Kind != "lib" {
		t.Fatalf("kind %s", f.m.Kind)
	}
	f.m = f.m.SelectRow(2).SelectRow(2) // the name row starts editing
	if !f.m.Editing() {
		t.Fatal("clicking the selected text row edits it")
	}
}
