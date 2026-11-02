package commands

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/actions/catalog"
)

// notInPalette are the catalogue actions the palette has no line of its own
// for, each with the way a person reaches it. A new action must be findable
// by its name or be listed here with a reason.
var notInPalette = map[string]string{
	"remove": "the Projects tab and the palette's Remove project", "swap": "palette: Swap", "upgrade": "palette: Upgrade",
	"release": "the Release tab", "sync": "Dependencies tab", "install": "the Install a package prompt", "uninstall": "the Remove a package prompt",
	"config": "Settings tab", "config set": "Settings tab", "config unset": "Settings tab", "architecture": "Settings tab",
	"plugin": "Settings tab", "plugin install": "Settings tab", "plugin set": "Settings tab", "plugin enable": "Settings tab", "plugin disable": "Settings tab",
	"recipe": "Settings tab", "recipe new": "the guided recipe form",
	"run": "run keys", "build": "run keys", "test": "run keys", "lint": "run keys", "typecheck": "run keys",
	"kind": "the Projects detail", "eject": "derived files flow", "derive": "palette: Update the derived files",
	"environments list": "no screen yet", "environments add": "no screen yet", "environments remove": "no screen yet", "tauri add": "Add project form", "tauri icon": "no screen yet",
	"deps up": "Dependencies tab", "up": "Dependencies tab", "ci": "palette: Regenerate CI pipelines", "ci generate": "palette: Regenerate CI pipelines",
	"pipeline": "no screen yet", "docker description": "no screen yet", "doctor": "palette: Doctor", "projects": "Projects tab", "affected": "palette: Affected",
	"info": "palette: Info", "explain": "run scope", "why-miss": "no screen yet", "report": "palette: Open the report",
	"cache status": "no screen yet", "cache prune": "no screen yet", "cache clear": "no screen yet", "cache verify": "no screen yet",
	"boundaries": "Doctor", "slices check": "Doctor", "mcp install": "no screen yet", "agent file": "palette: Write the agent file",
	"agent install": "Doctor fix", "divergences": "palette: Divergences", "try": "palette: Try a command", "bootstrap": "palette: Write the launcher",
	"adopt": "palette: Bring into MonoSmith", "new": "palette: New workspace", "add": "palette: Add project", "convert": "no screen yet", "undo": "palette: Undo", "graph": "palette: Graph",
	"dev": "own terminal", "tools": "palette: Install the missing tools", "copies": "command line only",
}

// TestEveryCatalogueActionIsFindableInThePalette keeps the rule that a person
// who types an action's own name in the palette finds it, or that the action
// states how else it is reached.
func TestEveryCatalogueActionIsFindableInThePalette(t *testing.T) {
	var text []string
	for _, e := range PaletteEntries(State{View: TabHome, InWorkspace: true, Project: "web", Projects: 4}) {
		text = append(text, strings.ToLower(e.Title+" "+e.Hint))
	}
	found := func(words string) bool {
		for _, s := range text {
			if strings.Contains(s, words) {
				return true
			}
		}
		return false
	}
	for _, name := range []string{"commit", "tools install", "compose"} {
		if !found(name) {
			t.Errorf("typing %q in the palette finds nothing", name)
		}
	}
	for _, a := range catalog.All() {
		if a.NotInTUI != "" || a.Safety == catalog.NeedsTerminal && a.Terminal != "" && notInPalette[a.ID()] != "" {
			continue
		}
		if _, excused := notInPalette[a.ID()]; excused {
			continue
		}
		if !found(a.ID()) {
			t.Errorf("%q is in the catalogue, but the palette has no entry for it and the action is not excused", a.ID())
		}
	}
}
