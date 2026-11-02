package commandline

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
	"github.com/spf13/cobra"
)

// visiblePaths are the command paths a user can run, without cobra's own.
func visiblePaths(c *cobra.Command, prefix []string) [][]string {
	var out [][]string
	for _, s := range c.Commands() {
		if s.Hidden || s.Name() == "help" || s.Name() == "completion" {
			continue
		}
		path := append(append([]string{}, prefix...), s.Name())
		out = append(out, path)
		out = append(out, visiblePaths(s, path)...)
	}
	return out
}

func TestEveryCommandIsAnActionAndEveryActionIsACommand(t *testing.T) {
	root := Root("0.0.0")
	exists := map[string]bool{}
	for _, path := range visiblePaths(root, nil) {
		exists[strings.Join(path, " ")] = true
		if _, ok := catalog.Find(path); !ok {
			t.Errorf("smith %s has no action: add one to libs/actions/catalog/catalog_config.go", strings.Join(path, " "))
		}
	}
	for _, a := range catalog.All() {
		if !exists[a.ID()] {
			t.Errorf("action %q names a command that is not there: remove or rename it in libs/actions/catalog/catalog_config.go", a.ID())
		}
	}
}

// A command that can change the workspace and takes --yes asks without a
// terminal, so the shell must append --yes after its own confirmation: that
// needs Asks on the action, or every run the shell starts fails with "nobody can
// confirm this change".
func TestEveryCommandThatTakesYesAndChangesTheWorkspaceIsMarkedAsAsking(t *testing.T) {
	root := Root("0.0.0")
	for _, path := range visiblePaths(root, nil) {
		cmd, _, err := root.Find(path)
		if err != nil || cmd == nil || cmd.Flags().Lookup("yes") == nil {
			continue
		}
		action, ok := catalog.Find(path)
		if !ok || action.Asks {
			continue
		}
		changes := action.Safety == catalog.Changes
		for _, v := range action.Variants {
			// A variant that already needs --yes is the confirmed form.
			changes = changes || (v.Safety == catalog.Changes && !slices.Contains(v.When, "--yes"))
		}
		if changes {
			t.Errorf("smith %s takes --yes and changes the workspace, but its action is not marked Asks: the shell would start it without --yes (libs/actions/catalog/catalog_config.go)", strings.Join(path, " "))
		}
	}
}

func TestEveryActionHasAFormOrAMenuEntryInTheShellOrIsExcused(t *testing.T) {
	shown := guide.PresentedWithForms(clicontract.DescribeCommands(Root("0.0.0")))
	for _, a := range catalog.All() {
		switch {
		case shown[a.ID()] && a.NotInTUI != "":
			t.Errorf("action %q is in the shell and also excused (%q): drop NotInTUI", a.ID(), a.NotInTUI)
		case !shown[a.ID()] && a.NotInTUI == "":
			t.Errorf("action %q has no menu entry or form in the shell, and a form could not be made for it: say why in NotInTUI (libs/actions/catalog/catalog_config.go)", a.ID())
		}
	}
}
