package releasing

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// packDistributions copies the wheel and the sdist the build task wrote for
// the version into params.PackDir, for a GitHub Release, and uploads nothing
// to any index. A dry run lists the ones the build directory holds now.
//
// Parameters: workspaceRoot is the workspace; params carries the project, the
// version and PackDir.
// Returns: the copied files, workspace-relative, in Assets.
// Errors: the build directory holds no distribution of the version, or a file
// cannot be copied.
func packDistributions(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	result := protocol.PublishResult{Registry: "GitHub Release"}
	built := filepath.Join(workspaceRoot, filepath.FromSlash(protocol.ProjectBuildDir(params.Project.Root)))
	found, _ := filepath.Glob(filepath.Join(built, "*"))
	sort.Strings(found)
	var files []string
	for _, f := range found {
		if isDistributionOf(filepath.Base(f), params.Version) {
			files = append(files, f)
		}
	}
	if params.DryRun {
		result.Skipped = "dry run"
		if len(files) == 0 {
			result.Notes = append(result.Notes, "the wheel and the sdist are made by the build task, so they are named once it has run; the release attaches both")
		}
		for _, f := range files {
			result.Assets = append(result.Assets, path.Join(params.PackDir, filepath.Base(f)))
		}
		return result, nil
	}
	if len(files) == 0 {
		return result, errors.New("no distribution for version " + params.Version + " in " + protocol.ProjectBuildDir(params.Project.Root) + "; build it first")
	}
	out := filepath.Join(workspaceRoot, filepath.FromSlash(params.PackDir))
	if err := os.MkdirAll(out, 0o755); err != nil {
		return result, err
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return result, err
		}
		if err := os.WriteFile(filepath.Join(out, filepath.Base(f)), raw, 0o644); err != nil {
			return result, err
		}
		result.Assets = append(result.Assets, path.Join(params.PackDir, filepath.Base(f)))
	}
	return result, nil
}
