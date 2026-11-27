package settingsview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/recipewizard"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// recipeWizardForm is the action the Write a recipe form presents: the row
// "Recipe creation wizard" opens that form.
const recipeWizardForm = "recipe new" // words:literal

// The two actions at the top of the Recipes section: write the recipe the
// workspace already follows (one row for each language its projects use), and
// open the form that writes any recipe.

// recipeActionRow is the row of a recipe action.
func (m Model) recipeActionRow(th theme.Theme, it item) string {
	w := 0
	for _, other := range m.items(Recipes) {
		if other.kind == kindWorkspaceRecipe || other.kind == kindRecipeWizard {
			w = max(w, pane.Width(m.recipeActionLabel(other))+2)
		}
	}
	label := pane.Cell(m.recipeActionLabel(it), w)
	switch {
	case it.kind == kindRecipeWizard:
		return "   " + th.Accent(label) + th.Muted("smith recipe new") // words:literal
	case it.idx < 0:
		return "   " + th.Muted(label) + th.Muted(words.Say("settingsview.row.recipe.workspace.none"))
	}
	return "   " + th.Accent(label) + th.Muted("smith recipe new --from-workspace") // words:literal
}

// recipeActionLabel is the name of a recipe action: the language is named
// when the workspace's projects use several.
func (m Model) recipeActionLabel(it item) string {
	switch {
	case it.kind == kindRecipeWizard:
		return words.Say("settingsview.row.recipe.wizard")
	case len(m.snap.WorkspaceRecipes) > 1:
		return words.Say("settingsview.row.recipe.workspace.language", "language", it.key)
	}
	return words.Say("settingsview.row.recipe.workspace")
}

// recipeActionDetail tells what a recipe action does: for the workspace's
// recipe, where it is read from and every level it fixes, so the person sees
// the recipe before Enter.
func (m Model) recipeActionDetail(th theme.Theme, it item) []string {
	if it.kind == kindRecipeWizard {
		return []string{th.Bold(words.Say("settingsview.row.recipe.wizard")), words.Say("settingsview.recipe.wizard.about"), th.Muted(words.Say("settingsview.form.about"))}
	}
	out := []string{th.Bold(m.recipeActionLabel(it))}
	if it.idx < 0 {
		return append(out, th.Failure(words.Say("settingsview.recipe.workspace.none", "reason", m.snap.WorkspaceRecipeErr)), words.Say("settingsview.recipe.workspace.about"))
	}
	r := m.snap.WorkspaceRecipes[it.idx]
	out = append(out, readFrom(r), th.Muted(words.Say("settingsview.recipe.workspace.fixes")))
	out = append(out, fixedLines(r)...)
	out = append(out, "", words.Say("settingsview.recipe.workspace.about"))
	if len(m.snap.WorkspaceRecipes) > 1 {
		var others []string
		for _, o := range m.snap.WorkspaceRecipes {
			if o.Language != r.Language {
				others = append(others, o.Language)
			}
		}
		out = append(out, "", th.Muted(words.Say("settingsview.recipe.workspace.others", "languages", strings.Join(others, ", "))))
	}
	return append(out, "", th.Muted(words.Say("settingsview.recipe.workspace.enter")))
}

// readFrom says which projects and which stack a recipe is read from.
func readFrom(r WorkspaceRecipe) string {
	stack := words.Say("settingsview.recipe.workspace.nostack")
	if r.Stack != "" {
		stack = words.Say("settingsview.recipe.workspace.stack", "stack", r.Stack)
	}
	return words.Count("settingsview.recipe.workspace.from", r.Projects, "language", r.Language, "stack", stack)
}

// fixedLines are the levels a recipe fixes, one a line: the level, then its plugins.
func fixedLines(r WorkspaceRecipe) []string {
	out := make([]string, 0, len(r.Fixed))
	for _, f := range r.Fixed {
		level, plugins, _ := strings.Cut(f, "=")
		out = append(out, "  "+pane.Cell(level, 12)+strings.ReplaceAll(plugins, ",", ", "))
	}
	return out
}

// newRecipeName asks the name of the recipe the workspace follows for one
// language; Enter shows the command line before it runs.
func newRecipeName(r WorkspaceRecipe) *editor {
	return &editor{
		kind:   editRecipeName,
		title:  words.Say("settingsview.recipe.workspace.title", "language", r.Language),
		about:  words.Say("settingsview.recipe.workspace.name"),
		plugin: r.Language,
		input:  newInput("", words.Say("settingsview.recipe.workspace.placeholder", "language", r.Language)),
		recipe: r,
	}
}

// recipeNameLines draws the name field, what the recipe fixes and the
// command line the name makes.
func (m Model) recipeNameLines(th theme.Theme, w int) []string {
	e := m.ed
	in := e.input
	in.Width = max(w-6, 4)
	lines := []string{"", " " + in.View(), "", " " + readFrom(e.recipe), " " + th.Muted(words.Say("settingsview.recipe.workspace.fixes"))}
	for _, l := range fixedLines(e.recipe) {
		lines = append(lines, " "+l)
	}
	argv, _ := workspaceRecipeArgv(e.plugin, e.input.Value())
	return append(lines, "", " "+th.Muted("smith "+strings.Join(argv, " "))) // words:literal
}

// workspaceRecipeArgv is the command line that writes the recipe the
// workspace follows for language under the typed name, or why the name is
// not one a recipe can take.
func workspaceRecipeArgv(language, typed string) ([]string, string) {
	name := strings.TrimSpace(typed)
	argv := []string{"recipe", "new", "--from-workspace", "--lang", language, "--name"}
	if name == "" {
		return append(argv, "<name>"), words.Say("settingsview.recipe.workspace.missing")
	}
	if _, err := recipewizard.PlanFromFlags(name, "", "", nil, nil); err != nil {
		return append(argv, name), err.Error()
	}
	return append(argv, name), ""
}
