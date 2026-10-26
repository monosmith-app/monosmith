package scaffolding

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-jvm/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Remove takes a removed project out of the build that listed it: its
// <module> in the root pom and its include in settings.gradle.kts, whichever
// the workspace has. A build that no longer names it is left alone.
func Remove(workspaceRoot string, params protocol.RemoveParams) (protocol.RemoveResult, error) {
	var res protocol.RemoveResult
	for _, edit := range []struct {
		file   string
		remove func(src string) (string, bool)
	}{
		{rootPom, func(src string) (string, bool) { return RemoveModule(src, params.Project.Root) }},
		{settings, func(src string) (string, bool) {
			return RemoveInclude(src, discovery.ProjectPath(params.Project.Root))
		}},
	} {
		file := filepath.Join(workspaceRoot, edit.file)
		src, err := os.ReadFile(file)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return res, err
		}
		out, changed := edit.remove(string(src))
		if !changed {
			continue
		}
		res.Files = append(res.Files, edit.file)
		if params.DryRun {
			continue
		}
		if err := os.WriteFile(file, []byte(out), 0o644); err != nil {
			return res, err
		}
	}
	return res, nil
}
