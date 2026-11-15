package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// VSCodeExtensionStack is the stack of a VS Code extension.
const VSCodeExtensionStack = "vscode-extension"

// BuildsLikeApp reports whether a project builds into the workspace's
// dist/<project dir> (ADR 0102): an app, or a VS Code extension, which ships
// a .vsix packed from that folder. Every other library or package builds
// inside its own folder.
//
// Parameters: kind is the project's kind; extension is true for a VS Code
// extension (package.json engines.vscode, or its stack).
// Returns: true for the workspace's dist.
func BuildsLikeApp(kind protocol.ProjectKind, extension bool) bool {
	return kind == protocol.ProjectApp || extension
}
