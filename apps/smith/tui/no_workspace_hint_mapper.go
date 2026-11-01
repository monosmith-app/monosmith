package tui

import "github.com/monosmith-app/monosmith/apps/smith/tui/commands"

// noWorkspaceHint says what to do outside a workspace: an empty folder can
// only start a new one, a folder with files can also adopt or convert them.
// A smith.toml that cannot be read is told as it is, with its line and column,
// and adoption is not offered: the folder already is a workspace.
func (s Shell) noWorkspaceHint() string {
	if s.problem != "" {
		return " smith.toml cannot be read: " + s.problem + ". Fix that line, then run smith again."
	}
	press := commands.ShortcutOf("file.new")
	if s.emptyFolder {
		return " Nothing here yet. Press " + press + " to create a workspace in this folder."
	}
	if s.adopter != nil {
		return " Not a smith workspace yet. Press " + commands.ShortcutOf("file.adopt") + " to bring this folder in, " + press + " for a new workspace"
	}
	return " No workspace here. Press " + press + " to create one, or run smith adopt to bring this folder in"
}
