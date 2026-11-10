package doctorfix

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/smithfix"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// previewLimit is the most bytes a preview copy of the workspace may hold.
const previewLimit = 64 << 20

// previewSkipped are the folders a preview copy leaves out: they are large and
// no plan reads them.
var previewSkipped = map[string]bool{
	".git": true, "node_modules": true, "target": true, "dist": true, "build": true,
	".venv": true, "venv": true, "bin": true, "obj": true, ".smith": true,
}

var errPreviewTooLarge = errors.New("the workspace is too large to copy for a preview")

// previewRoot gives the folder in which a fix's plan commands run. A fix that
// writes files and then runs smith (pin a toolchain, then smith ci generate) must
// be planned as if its files were already written. Planning in the real
// workspace would show the plan of the old files, or "up to date".
//
// Parameters: root is the workspace root; fix is the fix under review.
// Returns: the folder to plan in and a function that removes it; root itself
// (and a no-op) when the fix has no files or no step with a plan; an error when
// the copy could not be made, and root is returned with it.
func previewRoot(root string, fix protocol.Fix) (string, func(), error) {
	none := func() {}
	if len(fix.Files) == 0 {
		return root, none, nil
	}
	planned := false
	for _, s := range fix.Steps {
		if _, ok := smithfix.PlanArgv(s.Argv); ok {
			planned = true
		}
	}
	if !planned {
		return root, none, nil
	}
	tmp, err := os.MkdirTemp("", "smith-preview-")
	if err != nil {
		return root, none, err
	}
	cleanup := func() { _ = os.RemoveAll(tmp) }
	if err := copyWorkspace(root, tmp); err != nil {
		cleanup()
		return root, none, err
	}
	// smith opens only a git repository; the copy has no history, which a plan does not read.
	if err := exec.Command("git", "init", "-q", tmp).Run(); err != nil {
		cleanup()
		return root, none, err
	}
	if err := ApplyFiles(tmp, protocol.Fix{Files: fix.Files}); err != nil {
		cleanup()
		return root, none, err
	}
	return tmp, cleanup, nil
}

// copyWorkspace copies the files of a workspace, without the folders in
// previewSkipped, into dst.
func copyWorkspace(src, dst string) error {
	var total int64
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil || rel == "." {
			return err
		}
		if d.IsDir() {
			if previewSkipped[d.Name()] {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if total += int64(len(raw)); total > previewLimit {
			return errPreviewTooLarge
		}
		return os.WriteFile(filepath.Join(dst, rel), raw, 0o644)
	})
}
