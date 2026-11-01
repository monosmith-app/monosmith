package releasing

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WriteVersion records params.Version in the project's package.json and
// updates the ranges of workspace dependencies that are being released too.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	manifest := path.Join(params.Project.Root, "package.json")
	full := filepath.Join(workspaceRoot, filepath.FromSlash(manifest))
	src, err := os.ReadFile(full)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	out, err := SetVersions(src, params.Version, params.Dependencies)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	if !params.DryRun {
		if err := lineendings.WriteFile(full, out, 0o644); err != nil {
			return protocol.VersionResult{}, err
		}
	}
	return protocol.VersionResult{Files: []string{manifest}}, nil
}
