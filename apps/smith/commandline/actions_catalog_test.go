package commandline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/actions"
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
		if _, ok := actions.Find(path); !ok {
			t.Errorf("smith %s has no action: add one to apps/smith/actions/catalog_config.go", strings.Join(path, " "))
		}
	}
	for _, a := range actions.All() {
		if !exists[a.ID()] {
			t.Errorf("action %q names a command that is not there: remove or rename it in apps/smith/actions/catalog_config.go", a.ID())
		}
	}
}
