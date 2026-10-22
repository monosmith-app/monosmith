package tui

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
)

// presented is the id of every action a menu entry or a guided form stands for.
func presented(t *testing.T) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	add := func(where string, ids ...string) {
		for _, id := range ids {
			if _, ok := catalog.ByID(id); !ok {
				t.Errorf("%s presents %q, which is not an action", where, id)
			}
			out[id] = true
		}
	}
	for _, c := range commands.All() {
		add(c.ID, c.Presents...)
		if c.Task != "" {
			add(c.ID, c.Task)
		}
		if len(c.Argv) > 0 {
			a, ok := catalog.Find(c.Argv)
			if !ok {
				t.Errorf("%s runs %v, which is not an action", c.ID, c.Argv)
				continue
			}
			add(c.ID, a.ID())
		}
	}
	for _, g := range guide.Actions([]string{"node"}, []string{"node"}, nil) {
		if g.Presents != "" {
			add(g.Title, g.Presents)
		}
	}
	return out
}

func TestEveryActionIsPresentedOrExcusedInTheShell(t *testing.T) {
	shown := presented(t)
	for _, a := range catalog.All() {
		switch {
		case shown[a.ID()] && a.NotInTUI != "":
			t.Errorf("action %q is in the shell and also excused (%q): drop NotInTUI", a.ID(), a.NotInTUI)
		case !shown[a.ID()] && a.NotInTUI == "":
			t.Errorf("action %q is in no menu or guided form of the shell: add one, or say why not in NotInTUI (libs/actions/catalog/catalog_config.go)", a.ID())
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
