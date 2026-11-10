package releasing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// packGem builds the gem with gem build, exactly as publish does, into
// params.PackDir and pushes it to no gem server. A dry run only names the file.
//
// Parameters: workspaceRoot is the workspace; params carries the project, the
// version and PackDir.
// Returns: the .gem, workspace-relative, in Assets.
// Errors: the project has no gemspec, or gem cannot build it.
func packGem(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	name := params.Project.Name
	if len(params.Project.Exports) > 0 {
		name = params.Project.Exports[0]
	}
	file := path.Join(params.PackDir, name+"-"+GemVersion(params.Version)+".gem")
	result := protocol.PublishResult{Registry: "GitHub Release", Assets: []string{file}}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	spec, err := gemspec(workspaceRoot, params.Project)
	if err != nil {
		return result, err
	}
	if err := os.MkdirAll(filepath.Join(workspaceRoot, filepath.FromSlash(params.PackDir)), 0o755); err != nil {
		return result, err
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))
	if err := GemRunner(dir, []string{"gem", "build", spec, "--output", filepath.Join(workspaceRoot, filepath.FromSlash(file))}); err != nil {
		return result, err
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(file))); err != nil {
		return result, fmt.Errorf("gem build did not write %s", file)
	}
	return result, nil
}
