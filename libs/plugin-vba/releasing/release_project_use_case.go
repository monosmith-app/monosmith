package releasing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-vba/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WriteVersion records a release's version in the smithVersion comment of the
// project's main module, keeping the file's line endings (VBA exports CRLF).
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
	main, src := discovery.FindMain(names, func(name string) (string, bool) {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		return string(raw), err == nil
	})
	if main == "" {
		return protocol.VersionResult{}, fmt.Errorf("%s has no VBA module", params.Project.Root)
	}
	out, changed := discovery.SetField(src, "Version", params.Version)
	if !changed {
		return protocol.VersionResult{Files: []string{}}, nil
	}
	if !params.DryRun {
		if err := os.WriteFile(filepath.Join(dir, main), []byte(out), 0o644); err != nil {
			return protocol.VersionResult{}, err
		}
	}
	return protocol.VersionResult{Files: []string{path.Join(params.Project.Root, main)}}, nil
}

// Publish has no registry to push to: a project is released by its tag.
func Publish(protocol.PublishParams) (protocol.PublishResult, error) {
	return protocol.PublishResult{Skipped: "released by its tag"}, nil
}
