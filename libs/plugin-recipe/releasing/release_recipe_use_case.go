package releasing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-recipe/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// file finds a recipe project's recipe file, workspace-relative.
func file(workspaceRoot, projectRoot string) (rel string, err error) {
	name, ok := discovery.OwnFolder(filepath.Join(workspaceRoot, filepath.FromSlash(projectRoot)))
	if !ok {
		return "", fmt.Errorf("%s holds no single %s file", projectRoot, discovery.Manifest)
	}
	return path.Join(projectRoot, name), nil
}

// WriteVersion records a release's version in the recipe file.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	rel, err := file(workspaceRoot, params.Project.Root)
	if err != nil {
		return protocol.VersionResult{}, err
	}
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

// Publish has no registry: a recipe is released by its tag, and its file is
// the release's one asset, so a URL can point at a version of the recipe.
//
// The asset is a copy in dist/release/<name>/ with the release's version
// written in. The recipe in the repository keeps the version it was committed
// with (a release commits nothing), so only the copy says which release it is,
// and a resumed release makes the same copy.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	rel, err := file(workspaceRoot, params.Project.Root)
	if err != nil {
		return protocol.PublishResult{}, err
	}
	asset := path.Join("dist", "release", params.Project.Name, path.Base(rel))
	if !params.DryRun {
		src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
		if err != nil {
			return protocol.PublishResult{}, err
		}
		out, _ := SetVersion(string(src), params.Version)
		full := filepath.Join(workspaceRoot, filepath.FromSlash(asset))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return protocol.PublishResult{}, err
		}
		if err := os.WriteFile(full, []byte(out), 0o644); err != nil {
			return protocol.PublishResult{}, err
		}
	}
	return protocol.PublishResult{Skipped: "released by its tag", Assets: []string{asset}}, nil
}
