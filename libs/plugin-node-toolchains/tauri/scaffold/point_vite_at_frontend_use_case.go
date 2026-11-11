package scaffold

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/extras"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// viteConfigs are the Vite configuration files smith writes, in the order it
// looks for them.
var viteConfigs = []string{"vite.config.ts", "vite.config.mts", "vite.config.js", "vite.config.mjs"}

// PointViteAtFrontend makes the Vite build of a new Tauri project write into
// dist/tauri/<project dir>/frontend, the folder its tauri.conf.json takes the
// web app from (ADR 0092, owner decision 1). It changes only the outDir the
// vite plugin wrote ('<up>/dist/<project dir>'); a configuration without that
// value stays as it is, and smith doctor names the folder to use.
//
// Parameters: workspaceRoot is the absolute workspace folder; dir the
// workspace-relative project folder.
// Returns the workspace-relative path of the file it changed, or "" when it
// changed none, and an error when the file cannot be read or written.
func PointViteAtFrontend(workspaceRoot, dir string) (string, error) {
	shipped := "'" + protocol.BuildDirFromProject(dir) + "'"
	frontend := "'" + protocol.FromProject(dir, extras.FrontendDir(dir)) + "'"
	for _, name := range viteConfigs {
		file := filepath.Join(workspaceRoot, filepath.FromSlash(dir), name)
		raw, err := os.ReadFile(file)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		text := string(raw)
		if strings.Count(text, shipped) != 1 {
			return "", nil
		}
		if err := os.WriteFile(file, []byte(strings.Replace(text, shipped, frontend, 1)), 0o644); err != nil {
			return "", err
		}
		return strings.Trim(dir, "/") + "/" + name, nil
	}
	return "", nil
}
