package tui

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
)

func TestEveryPresentedActionIsAnActionAndIsNotAlsoExcused(t *testing.T) {
	for id := range Presented() {
		a, ok := catalog.ByID(id)
		switch {
		case !ok:
			t.Errorf("a menu entry or form presents %q, which is not an action", id)
		case a.NotInTUI != "":
			t.Errorf("action %q is in the shell and also excused (%q): drop NotInTUI", id, a.NotInTUI)
		}
	}
	for _, c := range commands.All() {
		// help is cobra's own command, which prints and has no action.
		if len(c.Argv) > 0 && c.Argv[0] != "help" {
			if _, ok := catalog.Find(c.Argv); !ok {
				t.Errorf("%s runs %v, which is not an action", c.ID, c.Argv)
			}
		}
	}
}

func TestEveryGuidedCommandNamesAGuidedForm(t *testing.T) {
	var titles []string
	for _, g := range guide.Actions([]string{"node"}, []string{"node"}, nil) {
		titles = append(titles, g.Title)
	}
	for _, c := range commands.All() {
		if c.Kind == commands.Guided && !slices.Contains(titles, c.Guide) {
			t.Errorf("%s opens the guided form %q, which does not exist", c.ID, c.Guide)
		}
	}
}
