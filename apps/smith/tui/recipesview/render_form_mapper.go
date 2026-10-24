package recipesview

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
)

// labelWidth is the width of the field names, so the values line up.
const labelWidth = 9

// View draws the form, what the recipe puts in each role, the command line
// the form builds and the recipes of the workspace, in w cells; rowOf maps a
// line to the form row drawn on it.
func (v Model) View(th theme.Theme, w int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	switch {
	case v.Catalog.Err != nil:
		return []string{th.Failure(v.Catalog.Err.Error())}, rowOf
	case !v.Loaded:
		return []string{th.Muted("Reading the workspace's recipes…")}, rowOf
	}
	if v.NoRecipes() {
		lines = append(lines, v.intro(th, w)...)
	}
	for i, r := range v.rows() {
		rowOf[len(lines)] = i
		lines = append(lines, v.field(th, i, r, w))
	}
	formEnd := len(lines)
	lines = append(lines, "")
	switch v.mode() {
	case NewRecipe:
		lines = append(lines, v.wizard(th)...)
	default:
		lines = append(lines, v.preview(th)...)
	}
	lines = append(lines, "")
	if argv, reason := v.Argv(); argv != nil {
		lines = append(lines, th.Title("Command")+"  "+th.Bold(commands.Line(argv)), th.Muted("c creates it"+confirmNote(argv)))
	} else {
		lines = append(lines, th.Title("Command")+"  "+th.Warning(reason))
	}
	lines = append(lines, "", th.Muted("up/down move · left/right or enter change a field · c creates"))
	lines = append(lines, v.recipes(th)...)
	if v.Problem != "" {
		lines = append(lines, th.Failure(v.Problem))
	}
	// Text below the form wraps; the form's own rows keep one line each.
	var out []string
	for i, l := range lines {
		if i < formEnd || pane.Width(l) <= w {
			out = append(out, pane.Fit(l, w))
			continue
		}
		for _, part := range pane.Wrap(l, w, 2) {
			out = append(out, pane.Fit(part, w))
		}
	}
	return out, rowOf
}

// intro explains what a recipe is to someone whose workspace has none.
func (v Model) intro(th theme.Theme, w int) []string {
	out := []string{th.Title("What is a recipe?")}
	for _, p := range []string{
		"A recipe is an optional team template. It says which plugins fill each role (language, lint, test, deploy) and which questions smith asks when someone adds a project, so every project starts the same way.",
		"You do not need one. Without a recipe smith uses the defaults of the project's stack or language. To add a project, use Add a project in the Guide.",
		"This workspace has no recipe file (*.recipe.toml) yet. To write the first one:",
	} {
		out = append(out, pane.Wrap("  "+p, w, 2)...)
		out = append(out, "")
	}
	return out
}

// confirmNote says what happens after c: the shell asks first, or hands over to a terminal.
func confirmNote(argv []string) string {
	if catalog.Classify(argv) == catalog.NeedsTerminal {
		return " in a terminal of its own, where the wizard asks its questions"
	}
	return " after you confirm"
}

// field draws one row of the form.
func (v Model) field(th theme.Theme, i int, r row, w int) string {
	pointer := "  "
	if i == v.Cursor {
		pointer = th.Accent(th.Icons.Pointer + " ")
	}
	label, value, note := "", "", ""
	switch r.id {
	case "mode":
		label, value = "Create", choose(th, modeNames[v.Mode])
	case "kind":
		label, value = "Kind", choose(th, v.Kind)
		if ref := v.Catalog.ByKind[v.Kind]; ref != "" {
			note = "made from " + ref
		} else {
			note = "no recipe: made by the stack or the language plugin"
		}
	case "name":
		label, value = "Name", v.Name
		if v.editing == "name" {
			value = v.input.View()
		} else if value == "" {
			value = th.Muted("(Enter to type)")
		}
	case "path":
		label, value = "Path", v.Path
		if v.editing == "path" {
			value = v.input.View()
		}
	case "recipe":
		label, value = "Recipe", choose(th, orNone(v.Recipe))
	case "choice":
		q := r.question
		label = q.Name
		value = q.Answer
		switch {
		case v.editing == "choice:"+q.Name:
			value = v.input.View()
		case len(q.Options) > 0:
			value = choose(th, orNone(q.Answer))
			note = strings.Join(q.Options, " · ")
		case value == "":
			value = th.Muted("(Enter to type)")
		}
		if q.Ask != "" {
			note = strings.TrimSpace(q.Ask + "  " + th.Muted(note))
		}
		if q.Answer == "" {
			value += " " + th.Warning("needs an answer")
		} else if !q.Given && q.Default != "" {
			value += " " + th.Muted("(default)")
		}
	}
	line := pointer + pane.Fit(label, labelWidth) + " " + value
	if note != "" && v.editing == "" {
		if avail := w - pane.Width(line) - 3; avail >= 8 {
			line += "   " + th.Muted(strings.TrimRight(pane.Cell(note, avail+1), " "))
		}
	}
	return line
}

// choose draws the value of a field that left and right change.
func choose(th theme.Theme, value string) string { return th.Bold("< " + value + " >") }

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}

// preview says what the recipe puts in each role, before anything is written.
func (v Model) preview(th theme.Theme) []string {
	ref := v.ref()
	if ref == "" {
		return []string{th.Muted("No recipe: nothing is chosen for you beyond the stack's or the language plugin's defaults.")}
	}
	d := v.Detail
	if d.Ref != ref {
		return []string{th.Muted("Reading " + ref + "…")}
	}
	title := "Recipe " + d.Name
	if d.Version != "" {
		title += " " + d.Version
	}
	out := []string{th.Title(title) + th.Muted("  "+ref)}
	switch {
	case d.Problem != "":
		return append(out, th.Failure("  "+d.Problem))
	case len(d.Roles) > 0:
		out = append(out, th.Muted("It puts, for these answers:"))
		keyW := 8
		for _, r := range d.Roles {
			keyW = max(keyW, pane.Width(r.Key))
		}
		for _, r := range d.Roles {
			out = append(out, "  "+pane.Fit(r.Key, keyW)+"  "+pane.Fit(strings.Join(r.Plugins, ", "), 22)+" "+th.Muted(r.Origin))
		}
		if len(d.Configures) > 0 {
			out = append(out, th.Muted("  and settings for: "+strings.Join(d.Configures, ", ")))
		}
	default:
		out = append(out, th.Warning("  Roles depend on "+strings.Join(d.Unanswered, ", ")+": answer it above."))
		out = append(out, offers(th, d.Offers)...)
	}
	if v.mode() == NewWorkspace {
		out = append(out, th.Muted("  Choices are asked when a project is added (smith add asks them)."))
	}
	return out
}

// offers lists what each role could be, for a recipe whose choices are not answered yet.
func offers(th theme.Theme, os []Offer) []string {
	var out []string
	for _, o := range os {
		var parts []string
		if len(o.Plugins) > 0 {
			parts = append(parts, strings.Join(o.Plugins, ", "))
		}
		var whens []string
		for when := range o.Alternatives {
			whens = append(whens, when)
		}
		sort.Strings(whens)
		for _, when := range whens {
			parts = append(parts, fmt.Sprintf("%s if %s", strings.Join(o.Alternatives[when], ", "), when))
		}
		out = append(out, "  "+th.Muted(o.Key+": ")+strings.Join(parts, " · "))
	}
	return out
}

// wizard explains the recipe wizard, which is smith's own and runs in a terminal.
func (v Model) wizard(th theme.Theme) []string {
	return []string{
		th.Title("The recipe wizard"),
		"  smith recipe new asks for the recipe's name, the choices it asks and the plugin",
		"  that fills each role, checks every combination of answers, and writes " + th.Bold(orNone(v.Path)) + ".",
		"  It asks on a terminal, so the shell steps aside for it and smith runs it.",
	}
}

// recipes lists the recipe files of the workspace.
func (v Model) recipes(th theme.Theme) []string {
	if len(v.Catalog.Recipes) == 0 {
		return nil
	}
	out := []string{"", th.Title(fmt.Sprintf("Recipes in this workspace (%d)", len(v.Catalog.Recipes)))}
	for _, l := range v.Catalog.Recipes {
		switch {
		case l.Error != "":
			out = append(out, "  "+th.Failure(th.Icons.Failed)+" "+l.Path+"  "+th.Failure(l.Error))
		default:
			line := "  " + th.Success(th.Icons.Passed) + " " + l.Path + "  " + th.Muted(strings.TrimSpace(l.Name+" "+l.Version))
			if l.Extends != "" {
				line += th.Muted("  extends " + l.Extends)
			}
			out = append(out, line)
		}
	}
	return out
}
