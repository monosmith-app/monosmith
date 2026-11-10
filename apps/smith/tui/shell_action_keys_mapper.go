package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
)

// actionCommands is the command each action of a detail stands for. The button
// of the action shows that command's key, so a click and the key are one thing.
var actionCommands = map[string]string{
	"x:install":        "projects.install-package",
	"x:uninstall":      "projects.remove-package",
	"x:explain":        "projects.explain",
	"x:why":            "projects.why",
	"x:graph":          "projects.graph",
	"x:derive-explain": "projects.derive-explain",
	"x:derive-apply":   "projects.derive-apply",
	"x:derive-adopt":   "projects.derive-adopt",
	"x:derive-eject":   "projects.derive-eject",
	"x:derive-replace": "projects.derive-replace",
	"x:rerun":          "run.again",
	"x:rerun-failed":   "run.again-failed",
	"x:copy-log":       "run.copy-log",
	"x:projects":       "view.projects",
	"x:doctor-fix":     "doctor.fix",
	"x:reload":         "view.reload",
}

// shortcutHere is the key of command id as the open view reads it: "" when that
// key does something else, or nothing, in this view, so a button never shows a
// key that does not work.
func (s Shell) shortcutHere(id string) string {
	c, ok := commands.ByID(id)
	if !ok || len(c.Keys) == 0 {
		return ""
	}
	if got, _, ok := commands.ByKey(c.Keys[0], s.state()); !ok || got.ID != id {
		return ""
	}
	return c.Shortcut
}

// withKeys gives each action of a detail that has no key yet the key of the
// command it stands for.
func (s Shell) withKeys(actions []detail.Action) []detail.Action {
	out := make([]detail.Action, len(actions))
	for i, a := range actions {
		if id, ok := actionCommands[a.ID]; ok && a.Shortcut == "" {
			a.Shortcut = s.shortcutHere(id)
		}
		out[i] = a
	}
	return out
}
