package scaffold

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// cliVersion is the @tauri-apps/cli a new project starts with; apiVersion is
// the @tauri-apps/api its web code imports.
const (
	cliVersion = "^2.0.0"
	apiVersion = "^2.0.0"
)

// Setup is the CLI and its script and, when the project has a web page to wrap
// (an index.html) and no src-tauri yet, the src-tauri that shows its build.
func Setup(workspaceRoot string, params protocol.GenerateParams) tooling.Setup {
	s := tooling.Setup{
		DevDependencies: map[string]string{"@tauri-apps/cli": cliVersion},
		Scripts:         map[string]string{"tauri": "tauri"},
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory), "index.html")); err == nil {
		s.Files = renderSrcTauri(params.Name, params.Workspace.Name, protocol.BuildDirFromProject(params.Directory))
		s.Fields = map[string]any{"dependencies": map[string]any{"@tauri-apps/api": apiVersion}}
	}
	return s
}
