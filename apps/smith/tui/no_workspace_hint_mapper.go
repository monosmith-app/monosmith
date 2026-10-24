package tui

// noWorkspaceHint says what to do outside a workspace: an empty folder can
// only start a new one, a folder with files can also adopt or convert them.
func (s Shell) noWorkspaceHint() string {
	if s.emptyFolder {
		return " Nothing here yet. File > New workspace… creates a workspace in this folder."
	}
	return " No workspace here. File > New workspace… (or smith adopt, smith convert)"
}
