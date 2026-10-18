package releasing

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Twine runs twine and returns its combined output; tests replace it.
var Twine = func(dir string, argv []string, env []string) (string, error) {
	out, err := toolrun.Output(dir, argv, env...)
	_, _ = toolrun.Log.Write([]byte(out))
	return out, err
}

// WriteVersion records the version in the project's pyproject.toml.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	manifest := path.Join(params.Project.Root, "pyproject.toml")
	full := filepath.Join(workspaceRoot, filepath.FromSlash(manifest))
	src, err := os.ReadFile(full)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	out, err := SetVersion(string(src), params.Version)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	if !params.DryRun {
		if err := os.WriteFile(full, []byte(out), 0o644); err != nil {
			return protocol.VersionResult{}, err
		}
	}
	return protocol.VersionResult{Files: []string{manifest}}, nil
}

// Publish uploads the build in the workspace's dist/<project dir> with
// twine, to the registry smith.toml names for it, else PyPI
// (TWINE_REPOSITORY_URL picks another index; PYPI_TOKEN, when set, is the
// API token).
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))
	target := targetOf(params.Project)
	result := protocol.PublishResult{Registry: target.Upload}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	built := filepath.Join(workspaceRoot, filepath.FromSlash(protocol.ProjectBuildDir(params.Project.Root)))
	dists, _ := filepath.Glob(filepath.Join(built, "*"))
	sort.Strings(dists)
	var files []string
	for _, d := range dists {
		if strings.Contains(filepath.Base(d), "-"+params.Version) {
			files = append(files, d)
		}
	}
	if len(files) == 0 {
		return result, errors.New("no distribution for version " + params.Version + " in " + protocol.ProjectBuildDir(params.Project.Root) + "; build it first")
	}
	project := params.Project.Name
	if len(params.Project.Exports) > 0 {
		project = params.Project.Exports[0]
	}
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = filepath.Base(f)
	}
	present := Uploaded(target.Index, project, names)
	var missing []string
	for i, f := range files {
		if !present[names[i]] {
			missing = append(missing, f)
		}
	}
	if len(missing) == 0 {
		result.Skipped = "already published"
		return result, nil
	}
	argv := append([]string{venv.Path(workspaceRoot), "-m", "twine", "upload", "--non-interactive", "--disable-progress-bar"}, target.Args...)
	argv = append(argv, missing...)
	if out, err := Twine(dir, argv, target.Env); err != nil {
		if strings.Contains(out, "File already exists") || strings.Contains(out, "409") {
			result.Skipped = "already published"
			return result, nil
		}
		return result, err
	}
	result.Published = true
	return result, nil
}
