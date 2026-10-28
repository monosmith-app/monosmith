package agentfiles

import (
	"os"
	"path/filepath"
	"strings"
)

// PresentClients says which clients the workspace or the machine has: Claude
// Code when the workspace has .claude or .mcp.json or the claude command is on
// the PATH, VS Code when it has .vscode or the code command is. With neither,
// both are present: a workspace gets the files of both. lookPath may be nil to
// look at the workspace only (smith doctor does, so its answer does not change
// from one machine to the next).
func PresentClients(root string, lookPath func(string) (string, error)) (claude, vscode bool) {
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(root, rel))
		return err == nil
	}
	on := func(cmd string) bool {
		if lookPath == nil {
			return false
		}
		_, err := lookPath(cmd)
		return err == nil
	}
	claude = exists(".claude") || exists(ClaudeMCPPath) || on("claude")
	vscode = exists(".vscode") || on("code")
	if !claude && !vscode {
		return true, true
	}
	return claude, vscode
}

// LauncherFor chooses how a new entry starts smith. A smith on the PATH is the
// usual case: every global way to install it (the install script, npm i -g,
// dotnet tool, uv, pipx, pip) puts it there. A smith run from the workspace's
// own node_modules (monosmith as a devDependency) is not on the PATH, and npx
// finds it. exe is the running binary.
func LauncherFor(root, exe string) Launcher {
	slash := func(p string) string { return strings.ToLower(strings.ReplaceAll(p, "\\", "/")) }
	if abs, err := filepath.Abs(root); err == nil && exe != "" &&
		strings.HasPrefix(slash(exe), slash(abs)+"/node_modules/") {
		return NpxLauncher
	}
	return DirectLauncher
}
