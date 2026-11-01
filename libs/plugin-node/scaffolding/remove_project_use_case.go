package scaffolding

import (
	"errors"
	"fmt"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Remove reinstalls at the workspace root once a project's directory is gone,
// so package-lock.json and node_modules forget the package.
func Remove(workspaceRoot string, params protocol.RemoveParams) (protocol.RemoveResult, error) {
	if _, err := os.Stat(filepath.Join(workspaceRoot, "package.json")); errors.Is(err, fs.ErrNotExist) {
		return protocol.RemoveResult{}, nil
	}
	var files []string
	if _, err := os.Stat(filepath.Join(workspaceRoot, "package-lock.json")); err == nil {
		files = append(files, "package-lock.json")
	}
	if params.DryRun {
		return protocol.RemoveResult{Files: files, Notes: []string{"would run npm install to update the lockfile"}}, nil
	}
	lockPath := filepath.Join(workspaceRoot, "package-lock.json")
	if lock, err := os.ReadFile(lockPath); err == nil {
		forgotten, changed, err := ForgetWorkspace(lock, params.Project.Root)
		if err != nil {
			return protocol.RemoveResult{}, fmt.Errorf("package-lock.json: %w", err)
		}
		if changed {
			if err := lineendings.WriteFile(lockPath, forgotten, 0o644); err != nil {
				return protocol.RemoveResult{}, err
			}
		}
	}
	if err := Installer(workspaceRoot); err != nil {
		return protocol.RemoveResult{}, fmt.Errorf("updating the lockfile: %w", err)
	}
	return protocol.RemoveResult{Files: files}, nil
}
