package commandline

import (
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
