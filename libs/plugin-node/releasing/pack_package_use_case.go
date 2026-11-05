package releasing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TarballName is the file npm pack writes for a package: the name without its
// "@", the scope's slash made a dash, then the version and ".tgz".
func TarballName(name, version string) string {
	name = strings.ReplaceAll(strings.TrimPrefix(name, "@"), "/", "-")
	return name + "-" + version + ".tgz"
}

// packPackage packs one package as publish would, with npm pack, into
// params.PackDir and publishes nothing: no registry is read or written. A dry
// run only names the file.
//
// Parameters: workspaceRoot is the workspace; params carries the project, the
// version and PackDir; name is the package name.
// Returns: the tarball, workspace-relative, in Assets.
// Errors: npm cannot pack, or does not write the file it should.
func packPackage(workspaceRoot string, params protocol.PublishParams, name string) (protocol.PublishResult, error) {
	file := path.Join(params.PackDir, TarballName(name, params.Version))
	result := protocol.PublishResult{Registry: "GitHub Release", Assets: []string{file}}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	staged, err := Stage(workspaceRoot, params.Project.Root)
	if err != nil {
		return result, err
	}
	out := filepath.Join(workspaceRoot, filepath.FromSlash(params.PackDir))
	if err := os.MkdirAll(out, 0o755); err != nil {
		return result, err
	}
	if err := npm.run(workspaceRoot, []string{"npm", "pack", staged, "--pack-destination", out}); err != nil {
		return result, fmt.Errorf("npm pack: %w", err)
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(file))); err != nil {
		return result, fmt.Errorf("npm pack did not write %s", file)
	}
	return result, nil
}
