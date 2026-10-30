package releasing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// rockspec is the project's rockspec, project-relative: the first one in its
// directory, as discovery reads it.
func rockspec(workspaceRoot string, p protocol.Project) (string, error) {
	found, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "*.rockspec"))
	if len(found) == 0 {
		return "", fmt.Errorf("%s has no rockspec", p.Root)
	}
	return filepath.Base(found[0]), nil
}

// WriteVersion records a release's version in the rockspec. The file keeps its
// name: LuaRocks expects it to carry the version, which smith doctor reports
// and offers to rename.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	file, err := rockspec(workspaceRoot, params.Project)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	rel := path.Join(params.Project.Root, file)
	full := filepath.Join(workspaceRoot, filepath.FromSlash(rel))
	src, err := os.ReadFile(full)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	out, changed := SetVersion(string(src), params.Version)
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

// Publish has no registry to push to: a rock is released by its tag.
func Publish(protocol.PublishParams) (protocol.PublishResult, error) {
	return protocol.PublishResult{Skipped: "released by its tag"}, nil
}
