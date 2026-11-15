package apptasks

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Folders are the folders a function or Electron app command works in.
type Folders struct {
	// Project is the project folder: where the command runs.
	Project string
	// Build is the folder the project builds into: dist/<project dir>.
	Build string
	// ElectronApp is the work folder that Electron runs and Forge packs:
	// dist/electron/<project dir>/app (ADR 0092).
	ElectronApp string
	// Name is the project name: the name of its folder.
	Name string
}

// AppFolders finds the folders of the project that a command runs in.
//
// Parameters: workspaceRoot is the absolute workspace folder; projectDir the
// absolute project folder, below it or the root itself.
// Returns the folders, or an error when projectDir is not inside the workspace.
func AppFolders(workspaceRoot, projectDir string) (Folders, error) {
	rel, err := filepath.Rel(workspaceRoot, projectDir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return Folders{}, fmt.Errorf("%s is not inside the workspace %s: run this command from a project folder", projectDir, workspaceRoot)
	}
	root := path.Clean(filepath.ToSlash(rel))
	build := protocol.BuildDir
	app := protocol.BuildDir + "/electron/app"
	if root != "." {
		build = protocol.ProjectBuildDir(root)
		app = protocol.BuildDir + "/electron/" + root + "/app"
	}
	return Folders{
		Project:     projectDir,
		Build:       filepath.Join(workspaceRoot, filepath.FromSlash(build)),
		ElectronApp: filepath.Join(workspaceRoot, filepath.FromSlash(app)),
		Name:        filepath.Base(projectDir),
	}, nil
}
