package releasing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-applescript/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WriteVersion records a release's version in the smithVersion comment of the
// project's main script.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	main := discovery.PickMain(names)
	if main == "" {
		return protocol.VersionResult{}, fmt.Errorf("%s has no .applescript source", params.Project.Root)
	}
	rel := path.Join(params.Project.Root, main)
	full := filepath.Join(dir, main)
	src, err := os.ReadFile(full)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	out, changed := discovery.SetField(string(src), "Version", params.Version)
	if !changed {
		return protocol.VersionResult{Files: []string{}}, nil
	}
	if !params.DryRun {
		if err := os.WriteFile(full, []byte(out), 0o644); err != nil {
			return protocol.VersionResult{}, err
		}
	}
	return protocol.VersionResult{Files: []string{rel}}, nil
}

// Publish has no registry to push to: a script is released by its tag.
func Publish(protocol.PublishParams) (protocol.PublishResult, error) {
	return protocol.PublishResult{Skipped: "released by its tag"}, nil
}
