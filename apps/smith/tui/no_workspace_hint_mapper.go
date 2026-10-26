package tui

import "github.com/monosmith-app/monosmith/apps/smith/tui/commands"

// noWorkspaceHint says what to do outside a workspace: an empty folder can
// only start a new one, a folder with files can also adopt or convert them.
func (s Shell) noWorkspaceHint() string {
	press := commands.ShortcutOf("file.new")
	if s.emptyFolder {
		return " Nothing here yet. Press " + press + " to create a workspace in this folder."
	}
	if s.adopter != nil {
		return " Not a smith workspace yet. Press " + commands.ShortcutOf("file.adopt") + " to bring this folder in, " + press + " for a new workspace"
	}
	return " No workspace here. Press " + press + " to create one (or smith adopt)"
}
