package mcpserver

import (
	"testing"

	"github.com/monosmith-app/monosmith/libs/actions/catalog"
)

func TestEveryToolRunsAnAction(t *testing.T) {
	for _, tool := range Tools {
		if _, ok := catalog.ByID(tool.Action); !ok {
			t.Errorf("%s runs the action %q, which does not exist (libs/actions/catalog/catalog_config.go)", tool.Name, tool.Action)
		}
	}
}

func TestEveryActionHasAToolOrIsExcused(t *testing.T) {
	served := map[string]bool{}
	for _, tool := range Tools {
		served[tool.Action] = true
	}
	for _, a := range catalog.All() {
		switch {
		case served[a.ID()] && a.NotInMCP != "":
			t.Errorf("action %q has an MCP tool and is also excused (%q): drop NotInMCP", a.ID(), a.NotInMCP)
		case !served[a.ID()] && a.NotInMCP == "":
			t.Errorf("action %q has no MCP tool: add one, or say why not in NotInMCP (libs/actions/catalog/catalog_config.go)", a.ID())
		}
	}
}
