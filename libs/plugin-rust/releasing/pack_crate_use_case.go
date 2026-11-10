package releasing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// packCrate writes the crate's .crate file with cargo package, the archive
// cargo publish would upload, into params.PackDir and publishes to no
// registry. The crate is not compiled (--no-verify): a crate that depends on a
// sibling released in the same run could not be verified, because the sibling
// is on no registry. A dry run only names the file.
//
// Parameters: workspaceRoot is the workspace; params carries the project, the
// version and PackDir; crate is the crate name.
// Returns: the .crate, workspace-relative, in Assets.
// Errors: cargo cannot package the crate, or writes no file of the expected name.
func packCrate(workspaceRoot string, params protocol.PublishParams, crate string) (protocol.PublishResult, error) {
	file := path.Join(params.PackDir, crate+"-"+params.Version+".crate")
	result := protocol.PublishResult{Registry: "GitHub Release", Assets: []string{file}}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	// cargo writes <target-dir>/package/<crate>-<version>.crate; the manifests
	// carry the release's versions while smith releases and are restored after,
	// so cargo must not refuse the dirty tree.
	target := filepath.Join(workspaceRoot, filepath.FromSlash(params.PackDir), "cargo-target")
	if err := CargoRunner(workspaceRoot, []string{"cargo", "package", "-p", crate, "--allow-dirty", "--no-verify", "--target-dir", target}); err != nil {
		return result, err
	}
	made, err := os.ReadFile(filepath.Join(target, "package", crate+"-"+params.Version+".crate"))
	if err != nil {
		return result, fmt.Errorf("cargo package did not write %s-%s.crate", crate, params.Version)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, filepath.FromSlash(file)), made, 0o644); err != nil {
		return result, err
	}
	_ = os.RemoveAll(target)
	return result, nil
}
