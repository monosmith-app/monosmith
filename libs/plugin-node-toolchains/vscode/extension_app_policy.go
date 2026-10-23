package vscode

import (
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node/releasing"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// isExtensionApp reports whether a project is an app whose manifest is a VS Code extension.
func isExtensionApp(workspaceRoot string, p protocol.Project) bool {
	if p.Kind != protocol.ProjectApp {
		return false
	}
	_, ok := releasing.Extension(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root)))
	return ok
}
