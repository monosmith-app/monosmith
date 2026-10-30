package releasing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-objc/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WriteVersion records a release's version in the project's podspec: the
// first one in its directory, as discovery reads it. A podspec whose version is
// computed has no literal to write and is left alone.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	found, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root), discovery.ManifestPattern))
	if len(found) == 0 {
		return protocol.VersionResult{}, fmt.Errorf("%s has no podspec", params.Project.Root)
	}
	rel := path.Join(params.Project.Root, filepath.Base(found[0]))
	src, err := os.ReadFile(found[0])
	if err != nil {
		return protocol.VersionResult{}, err
	}
	out, changed := discovery.SetVersion(string(src), params.Version)
	if !changed {
		return protocol.VersionResult{Files: []string{}}, nil
	}
	if !params.DryRun {
		if err := os.WriteFile(found[0], []byte(out), 0o644); err != nil {
			return protocol.VersionResult{}, err
		}
	}
	return protocol.VersionResult{Files: []string{rel}}, nil
}

// Publish has nothing to push: a pod's source is found by the git tag the
// release makes (s.source's :tag). Pushing the podspec to CocoaPods trunk
// (pod trunk push) is not done by smith.
func Publish(protocol.PublishParams) (protocol.PublishResult, error) {
	return protocol.PublishResult{Skipped: "released by its tag"}, nil
}
