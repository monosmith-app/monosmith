package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"

	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// noWorkspaceHint says what to do outside a workspace: an empty folder can
// only start a new one, a folder with files can also adopt or convert them.
// A smith.toml that cannot be read is told as it is, with its line and column,
// and adoption is not offered: the folder already is a workspace.
func (s Shell) noWorkspaceHint() string {
	if s.problem != "" && s.repair != nil {
		return words.Say("tui.no_workspace_hint.smith_toml_cannot_be_read_press", "problem", s.problem)
	}
	if s.problem != "" {
		return words.Say("tui.no_workspace_hint.smith_toml_cannot_be_read_fix", "problem", s.problem)
	}
	press := commands.ShortcutOf("file.new")
	if s.emptyFolder {
		return words.Say("tui.no_workspace_hint.nothing_here_yet_press_to_create", "press", press)
	}
	if s.adopter != nil {
		return words.Say("tui.no_workspace_hint.not_a_smith_workspace_yet_press", "shortcutOf", commands.ShortcutOf("file.adopt"), "press", press)
	}
	return words.Say("tui.no_workspace_hint.no_workspace_here_press_to_create", "press", press)
}
