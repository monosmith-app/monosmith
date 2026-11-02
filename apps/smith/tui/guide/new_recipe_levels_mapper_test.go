package guide

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/monosmith-app/monosmith/apps/smith/recipewizard"
)

func recipeCatalog() recipewizard.LevelCatalog {
	offer := func(names ...string) []recipewizard.PluginOffer {
		var out []recipewizard.PluginOffer
		for _, n := range names {
			out = append(out, recipewizard.PluginOffer{Name: n, Description: "about " + n})
		}
		return out
	}
	return recipewizard.LevelCatalog{
		Levels: []recipewizard.LevelOffer{
			{Level: "language", Plugins: offer("node", "python")},
			{Level: "lint", Plugins: offer("eslint", "ruff")},
			{Level: "structure", Additive: true, Plugins: offer("layers", "slices")},
		},
		Fits: map[string]map[string][]string{
			"node":   {"lint": {"eslint"}, "structure": {"layers", "slices"}},
			"python": {"lint": {"ruff"}, "structure": {"layers", "slices"}},
		},
	}
}

func recipeForm() Model {
	base := Action{Title: "Write a recipe", Steps: NewRecipeSteps(), Argv: NewRecipeArgv}
	m, _ := NewModel([]Action{RecipeAction(base, recipeCatalog())}).Pick("Write a recipe")
	return m
}

func argvOf(m tea.Model) string { return strings.Join(m.(Model).Argv, " ") }

func TestRecipeFormFixesPluginsFromPickListsAndTickBoxes(t *testing.T) {
	// name, version, extends; language python; lint ruff; structure layers + slices.
	m := keys(recipeForm(), "team", "enter", "enter", "enter", "down", "down", "enter", "down", "enter", "enter", "space", "down", "space", "enter")
	want := "recipe new --name team --fixed language=python --fixed lint=ruff --fixed structure=layers,slices"
	if got := argvOf(m); got != want {
		t.Fatalf("got %q", got)
	}
}

func TestRecipeFormStartsOnLeaveUnsetAndTicksNothing(t *testing.T) {
	m := keys(recipeForm(), "team", "enter", "enter", "enter")
	if view := m.View(); !strings.Contains(view, "Leave unset (default)") || !strings.Contains(view, "about python") {
		t.Fatalf("the language question lists unset and the plugins with their description:\n%s", view)
	}
	// Enter on every question: nothing is fixed, and the command is the shortest.
	m = keys(m, "enter", "enter", "enter", "enter")
	if got := argvOf(m); got != "recipe new --name team" {
		t.Fatalf("got %q", got)
	}
}

func TestRecipeFormOffersOnlyPluginsThatFitTheLanguage(t *testing.T) {
	m := keys(recipeForm(), "team", "enter", "enter", "enter", "down", "enter")
	view := m.View()
	if !strings.Contains(view, "eslint") || strings.Contains(view, "ruff") {
		t.Fatalf("node was fixed, so only eslint fits:\n%s", view)
	}
}

func TestRecipeFormUntickingTakesAPluginBack(t *testing.T) {
	m := keys(recipeForm(), "team", "enter", "enter", "enter", "down", "enter", "enter", "enter", "space")
	if !strings.Contains(m.View(), "[x] layers") {
		t.Fatalf("space ticks:\n%s", m.View())
	}
	m = keys(m, "space")
	if !strings.Contains(m.View(), "[ ] layers") {
		t.Fatalf("space again unticks:\n%s", m.View())
	}
	m = keys(m, "enter")
	if got := argvOf(m); got != "recipe new --name team --fixed language=node" {
		t.Fatalf("got %q", got)
	}
}

func TestRecipeLevelChecksRefuseWhatTheCommandRefuses(t *testing.T) {
	a := map[string][]string{"name": {"team"}}
	check := recipeLevelCheck(recipeLevelKey("language", ""))
	if got := check(a, "python,node"); !strings.Contains(got, "takes one plugin") {
		t.Errorf("two language plugins were accepted: %q", got)
	}
	if got := check(a, "python"); got != "" {
		t.Errorf("one language plugin was refused: %q", got)
	}
	if got := recipeLevelCheck(recipeLevelKey("structure", "python"))(a, "layers,slices"); got != "" {
		t.Errorf("two structure plugins were refused: %q", got)
	}
}

func TestAPickThatIsRefusedDoesNotGoOn(t *testing.T) {
	refuse := func(map[string][]string, string) string { return "layers and slices cannot meet" }
	for _, multi := range []bool{false, true} {
		step := Step{Key: "k", Prompt: "Which?", Multi: multi, Check: refuse, Options: []Option{{Label: "a", Value: "a"}, {Label: "b", Value: "b"}}}
		m, _ := NewModel([]Action{{Title: "T", Steps: []Step{step}}}).Pick("T")
		m = keys(m, "space", "enter").(Model)
		if m.Done() || !strings.Contains(m.View(), "cannot meet") {
			t.Fatalf("multi=%v: a refused pick went on or hid the reason:\n%s", multi, m.View())
		}
	}
}

func TestParseRecipeLevelsReadsTheCommandsJSON(t *testing.T) {
	cat, err := ParseRecipeLevels(`{"levels":[{"level":"lint","additive":false,"plugins":[{"name":"ruff","description":"d"}]}],"fits":{"python":{"lint":["ruff"]}}}`)
	if err != nil || len(cat.Levels) != 1 || !reflect.DeepEqual(cat.Fits["python"]["lint"], []string{"ruff"}) {
		t.Fatalf("cat=%+v err=%v", cat, err)
	}
	if _, err := ParseRecipeLevels("not json"); err == nil {
		t.Fatal("text that is not JSON was accepted")
	}
}

func TestNoLevelsKeepsThePlainForm(t *testing.T) {
	base := Action{Title: "Write a recipe", Steps: NewRecipeSteps()}
	if got := RecipeAction(base, recipewizard.LevelCatalog{}); len(got.Steps) != len(base.Steps) {
		t.Fatal("an empty catalog must leave the plain form")
	}
}

func TestRecipeFormLetsTheUsersChooseALevelAndWritesChosen(t *testing.T) {
	// name, version, extends; language: the last answer, "chosen by the people who use it".
	m := keys(recipeForm(), "team", "enter", "enter", "enter", "down", "down", "down", "enter")
	if view := m.View(); !strings.Contains(view, "question do the users read") {
		t.Fatalf("the question the users read is asked next:\n%s", view)
	}
	// the question, the two plugins they can end up with, the answers (Enter keeps the plugin names).
	m = keys(m, "Which language?", "enter", "space", "down", "space", "enter")
	if view := m.View(); !strings.Contains(view, "node, python") {
		t.Fatalf("the answers start on the plugin names, to edit:\n%s", view)
	}
	m = keys(m, "enter", "enter", "enter", "enter", "enter")
	want := "recipe new --name team --chosen language=Which language?:node=node,python=python"
	if got := argvOf(m); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRecipeFormLetsTheUsersChooseALevelThatTakesManyPlugins(t *testing.T) {
	// language node, lint unset; structure: chosen, a question, then answers where one joins two plugins.
	m := keys(recipeForm(), "team", "enter", "enter", "enter", "down", "enter", "enter", "down", "enter", "Which layout?", "enter")
	if view := m.View(); !strings.Contains(view, "layers=layers, slices=slices") {
		t.Fatalf("each plugin starts as an answer of its own, to edit:\n%s", view)
	}
	want := "recipe new --name team --fixed language=node --chosen structure=Which layout?:layers=layers,slices=slices"
	if got := argvOf(keys(m, "enter")); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRecipeGroupsRefuseWhatTheCommandRefuses(t *testing.T) {
	offered := []string{"layers", "slices"}
	for text, want := range map[string]string{
		"one=layers":          "at least two answers",
		"a=layers, a=slices":  "two answers are named",
		"a=layers, b=hexagon": "not a plugin of this level",
		"a=layers, b":         "answer=plugin",
	} {
		if _, got := parseGroups(text, offered); !strings.Contains(got, want) {
			t.Errorf("%q: got %q, want %q", text, got, want)
		}
	}
	pairs, problem := parseGroups("simple=layers, both=layers+slices", offered)
	if problem != "" || strings.Join(pairs, ",") != "simple=layers,both=layers+slices" {
		t.Errorf("pairs=%v problem=%q", pairs, problem)
	}
}

func TestRecipeFormRefusesAChoiceWithOneAnswer(t *testing.T) {
	m := keys(recipeForm(), "team", "enter", "enter", "enter", "down", "down", "down", "enter", "Which language?", "enter", "space", "enter")
	if !strings.Contains(m.View(), "at least two plugins") {
		t.Fatalf("one plugin is a fixed plugin, not a choice:\n%s", m.View())
	}
}
