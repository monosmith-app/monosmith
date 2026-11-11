package electronapp

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WorkDir is the work folder of a project's Electron packing, workspace-
// relative: dist/electron/<project dir> (ADR 0092, owner decisions 2 and 5).
//
// Parameters: root is the workspace-relative project folder.
// Returns the folder, workspace-relative.
func WorkDir(root string) string {
	return protocol.BuildDir + "/electron/" + strings.Trim(root, "/")
}

// WantedOutDir is where the tool of a setup should pack, workspace-relative:
// the work folder's out/ for Electron Forge (as the app smith makes), the
// work folder itself for electron-builder, which also writes its own work
// files there (*-unpacked/, builder-debug.yml).
//
// Parameters: root is the workspace-relative project folder; tool the setup.
// Returns the folder, workspace-relative.
func WantedOutDir(root string, tool Tool) string {
	if tool == Builder {
		return WorkDir(root)
	}
	return WorkDir(root) + "/out"
}
