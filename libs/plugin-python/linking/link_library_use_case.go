package linking

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-python/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-python/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/pelletier/go-toml/v2"
)

const (
	table = "project"
	key   = "dependencies"
)

// Distribution reads the name a library is installed and listed under (the
// [project] name of its pyproject.toml, shop-core-py), which is not the name
// it is imported by (shop_core_py).
func Distribution(workspaceRoot string, library protocol.Project) (string, error) {
	manifest := library.Root + "/pyproject.toml"
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(manifest)))
	if err != nil {
		return "", fmt.Errorf("%s: %w", manifest, err)
	}
	var py discovery.PyProject
	if err := toml.Unmarshal(raw, &py); err != nil {
		return "", fmt.Errorf("%s: %w", manifest, err)
	}
	if py.Project.Name == "" {
		return "", fmt.Errorf("%s has no [project] name, so %s cannot depend on it: give the library a name", manifest, library.Name)
	}
	return py.Project.Name, nil
}

// AddRequirement lists distribution (with spec, when given) in the
// [project] dependencies of a pyproject.toml text. It returns the new text and
// false when the file already lists the distribution.
func AddRequirement(src, distribution, spec string) (string, bool) {
	l := distribution + spec
	if has(src, distribution) {
		return src, false
	}
	return dependencies.AddRequirements(src, table, key, []string{l}), true
}

// RemoveRequirement takes distribution out of the [project] dependencies. It
// returns the new text and false when the file did not list it.
func RemoveRequirement(src, distribution string) (string, bool) {
	out, removed := dependencies.RemoveRequirements(src, table, key, []string{distribution})
	return out, len(removed) > 0
}

func has(src, distribution string) bool {
	var py discovery.PyProject
	if toml.Unmarshal([]byte(src), &py) != nil {
		return false
	}
	want := dependencies.RequirementName(distribution)
	for _, r := range py.Project.Dependencies {
		if dependencies.RequirementName(r) == want {
			return true
		}
	}
	return false
}

// Link lists the library's distribution name in the project's pyproject.toml
// and installs the library into the shared venv before the project, so pip
// finds the requirement already installed there and never asks PyPI for it.
// With params.Plan it writes nothing.
func Link(workspaceRoot string, params protocol.LinkParams) (protocol.LinkResult, error) {
	manifest := params.Project.Root + "/pyproject.toml"
	path := filepath.Join(workspaceRoot, filepath.FromSlash(manifest))
	raw, err := os.ReadFile(path)
	if err != nil {
		return protocol.LinkResult{}, fmt.Errorf("%s: %w", manifest, err)
	}
	distribution, err := Distribution(workspaceRoot, params.Library)
	if err != nil {
		return protocol.LinkResult{}, err
	}
	res := protocol.LinkResult{Manifest: manifest, Line: fmt.Sprintf("%q in dependencies", distribution+params.Spec)}
	if params.Plan {
		return res, nil
	}
	updated, changed := AddRequirement(string(raw), distribution, params.Spec)
	if !changed {
		res.Notes = []string{params.Project.Name + " already lists " + distribution}
		return res, nil
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return protocol.LinkResult{}, err
	}
	created, err := venv.Ensure(workspaceRoot)
	if err != nil {
		return protocol.LinkResult{}, err
	}
	// The library first: it is then installed in the venv, which is where pip
	// looks before it asks an index.
	for _, root := range []string{params.Library.Root, params.Project.Root} {
		if err := venv.InstallEditable(workspaceRoot, root); err != nil {
			return protocol.LinkResult{}, fmt.Errorf("%w\nIf pip looked for %s on PyPI, the library is not installed in the venv: run smith install", err, distribution)
		}
	}
	res.Files = append(created, manifest)
	return res, nil
}

// Unlink takes the library's distribution out of the project's pyproject.toml.
// The library stays installed in the venv, which every project shares. When the
// file did not list it, the result has no files.
func Unlink(workspaceRoot string, params protocol.UnlinkParams) (protocol.DependencyResult, error) {
	manifest := params.Project.Root + "/pyproject.toml"
	path := filepath.Join(workspaceRoot, filepath.FromSlash(manifest))
	raw, err := os.ReadFile(path)
	if err != nil {
		return protocol.DependencyResult{}, fmt.Errorf("%s: %w", manifest, err)
	}
	distribution, err := Distribution(workspaceRoot, params.Library)
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	updated, changed := RemoveRequirement(string(raw), distribution)
	if !changed {
		return protocol.DependencyResult{}, nil
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{Files: []string{manifest}, Notes: []string{"the library stays installed in .venv, which every project shares"}}, nil
}
