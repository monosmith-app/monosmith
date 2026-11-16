package manifestfixes

import (
	"path"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// BuildInPlaceFix moves the build of a library or package that compiles
// outside its folder (../../dist/<project dir>) back into its own dist (ADR
// 0102). The tsconfig or build script names ./dist, the package.json entries
// that led to the old folder (main, types, exports, scripts) follow it, and an
// outputs override in smith.toml for the old folder is removed. It returns nil
// when there is nothing to edit. Doctor asks before it applies it.
//
// Parameters: workspaceRoot is the absolute workspace folder; p the project;
// configName the tsconfig that sets oldOutDir ("" when the build script's
// flag does); oldOutDir the folder the build writes now, relative to the project.
// Returns: the fix, or nil.
func BuildInPlaceFix(workspaceRoot string, p protocol.Project, configName, oldOutDir string) *protocol.Fix {
	move := moveBuildFix(workspaceRoot, p, configName, oldOutDir, "./dist", true)
	if move == nil {
		return nil
	}
	move.Description = "Build " + p.Name + " inside its own folder (" + path.Join(p.Root, "dist") + ")"
	move.AppliesTo = nil
	if step, ok := dropDeclaredOutputs(workspaceRoot, p, cleanDir(oldOutDir), path.Join(p.Root, "dist")); ok {
		move.Steps = append(move.Steps, step)
	}
	return move
}

// InsideProject reports whether a workspace-relative folder is in a project's
// folder (and not the folder itself).
func InsideProject(root, folder string) bool {
	return root != "." && root != "" && strings.HasPrefix(folder, strings.TrimSuffix(root, "/")+"/")
}
