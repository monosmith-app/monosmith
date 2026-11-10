package releasing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// packNupkg packs the project with dotnet pack, exactly as publish does, into
// params.PackDir and pushes it to no feed. A dry run only names the file.
//
// Parameters: workspaceRoot is the workspace; params carries the project, the
// version and PackDir.
// Returns: the .nupkg, workspace-relative, in Assets.
// Errors: the project has no project file, dotnet cannot pack, or it writes
// no .nupkg of the expected name.
func packNupkg(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	id := params.Project.Name
	if len(params.Project.Exports) > 0 {
		id = params.Project.Exports[0]
	}
	file := path.Join(params.PackDir, id+"."+params.Version+".nupkg")
	result := protocol.PublishResult{Registry: "GitHub Release", Assets: []string{file}}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	project, err := projectFile(workspaceRoot, params.Project)
	if err != nil {
		return result, err
	}
	out := filepath.Join(workspaceRoot, filepath.FromSlash(params.PackDir))
	if err := os.MkdirAll(out, 0o755); err != nil {
		return result, err
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))
	if err := DotnetRunner(dir, []string{"dotnet", "pack", project, "--nologo", "-c", "Release", "-o", out, "-p:Version=" + params.Version}); err != nil {
		return result, err
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(file))); err != nil {
		return result, fmt.Errorf("dotnet pack did not write %s", file)
	}
	return result, nil
}
