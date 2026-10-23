package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
)

// Presented are the ids of the actions a menu entry or a form written by hand
// stands for.
func Presented() map[string]bool {
	out := map[string]bool{}
	for _, c := range commands.All() {
		out = addAll(out, c.Presents...)
		if c.Task != "" {
			out[c.Task] = true
		}
		if a, ok := catalog.Find(c.Argv); ok && len(c.Argv) > 0 {
			out[a.ID()] = true
		}
	}
	for _, g := range guide.Actions(nil, nil, nil) {
		if g.Presents != "" {
			out[g.Presents] = true
		}
	}
	return out
}

func addAll(set map[string]bool, ids ...string) map[string]bool {
	for _, id := range ids {
		set[id] = true
	}
	return set
}

// GeneratedForms are the forms of the actions the shell shows that no menu
// entry or form written by hand stands for, built from the commands'
// descriptions: a command added to smith has a form in the shell without
// anyone remembering to write one. An action left out on purpose (its
// NotInTUI says why) has none.
func GeneratedForms(described []clicontract.CommandDescription) []guide.Action {
	shown := Presented()
	var forms []guide.Action
	for _, a := range catalog.All() {
		if a.NotInTUI != "" || shown[a.ID()] {
			continue
		}
		if c, ok := describedAt(described, a.Command); ok {
			forms = append(forms, guide.FormFor(c, a.Command, a.ID()))
		}
	}
	return forms
}

// PresentedWithForms is Presented and the actions GeneratedForms gives a form.
func PresentedWithForms(described []clicontract.CommandDescription) map[string]bool {
	out := Presented()
	for _, f := range GeneratedForms(described) {
		out[f.Presents] = true
	}
	return out
}

// describedAt finds a command's description by its words: a top-level command,
// or a subcommand under its parent.
func describedAt(described []clicontract.CommandDescription, path []string) (clicontract.CommandDescription, bool) {
	for _, c := range described {
		if len(path) == 0 || c.Name != path[0] {
			continue
		}
		if len(path) == 1 {
			return c, true
		}
		return describedAt(c.Subcommands, path[1:])
	}
	return clicontract.CommandDescription{}, false
}
