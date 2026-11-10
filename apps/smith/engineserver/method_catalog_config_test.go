package engineserver

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/mcpserver"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
)

// served is the id of every action a method of the engine runs.
func served(t *testing.T) map[string]bool {
	t.Helper()
	out := map[string]bool{"run": true}
	toolAction := func(method, tool string) {
		for _, tl := range mcpserver.Tools {
			if tl.Name == tool {
				out[tl.Action] = true
				return
			}
		}
		t.Errorf("method %s maps to the tool %q, which does not exist", method, tool)
	}
	for method, tool := range queryMethods {
		toolAction(method, tool)
	}
	for method, tool := range commandMethods {
		toolAction(method, tool)
	}
	for method := range serverCommands {
		if _, ok := serverActions[method]; !ok {
			t.Errorf("method %s has no action in serverActions", method)
		}
	}
	for method := range serverMethods {
		if _, ok := serverActions[method]; !ok {
			t.Errorf("method %s has no action in serverActions", method)
		}
	}
	for method, id := range serverActions {
		if _, ok := catalog.ByID(id); !ok {
			t.Errorf("method %s runs the action %q, which does not exist", method, id)
		}
		out[id] = true
	}
	return out
}

func TestEveryActionHasAMethodOrIsExcused(t *testing.T) {
	shown := served(t)
	for _, a := range catalog.All() {
		switch {
		case shown[a.ID()] && a.NotInServe != "":
			t.Errorf("action %q has an engine method and is also excused (%q): drop NotInServe", a.ID(), a.NotInServe)
		case !shown[a.ID()] && a.NotInServe == "":
			t.Errorf("action %q has no engine method: add one, or say why not in NotInServe (libs/actions/catalog/catalog_config.go)", a.ID())
		}
	}
}
