package linking

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node/dependencies"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var quiet = []string{"--no-audit", "--no-fund"}

// Link makes the project depend on the library by its package name in the
// dependencies of its package.json, with the range the workspace already uses
// for such links, the one the user chose, or the default, and then runs npm
// install so the workspace's lock file and links follow. With params.Plan it
// writes nothing and answers what it would write and the choices of the range.
func Link(workspaceRoot string, params protocol.LinkParams) (protocol.LinkResult, error) {
	manifestPath := params.Project.Root + "/package.json"
	file := filepath.Join(workspaceRoot, filepath.FromSlash(manifestPath))
	raw, err := os.ReadFile(file)
	if err != nil {
		return protocol.LinkResult{}, fmt.Errorf("%s: %w", manifestPath, err)
	}
	if len(params.Library.Exports) == 0 {
		return protocol.LinkResult{}, fmt.Errorf("%s/package.json has no \"name\", so %s cannot depend on it: give the package a name", params.Library.Root, params.Project.Name)
	}
	name := params.Library.Exports[0]
	spec, options := chooseSpec(workspaceRoot, params)
	res := protocol.LinkResult{Manifest: manifestPath, Line: fmt.Sprintf("%q: %q", name, spec), Options: options}
	if params.Plan {
		return res, nil
	}
	updated, changed, err := AddDependency(raw, name, spec)
	if err != nil {
		return protocol.LinkResult{}, fmt.Errorf("%s: %w", manifestPath, err)
	}
	if !changed {
		res.Notes = []string{params.Project.Name + " already depends on " + name + " with " + spec}
		return res, nil
	}
	if err := os.WriteFile(file, updated, 0o644); err != nil {
		return protocol.LinkResult{}, err
	}
	notes, err := dependencies.FoldedNpmRunner(workspaceRoot, append([]string{"npm", "install"}, quiet...))
	if err != nil {
		return protocol.LinkResult{}, err
	}
	res.Files = []string{manifestPath, "package-lock.json"}
	res.Notes = notes
	res.Options = nil
	return res, nil
}

// chooseSpec picks the range to write: the user's, else the one every existing
// link in the workspace uses, else the default of two real choices.
func chooseSpec(workspaceRoot string, params protocol.LinkParams) (string, []protocol.LinkOption) {
	if params.Spec != "" {
		return params.Spec, nil
	}
	if used := InternalSpecs(workspaceRoot); len(used) == 1 {
		return used[0], nil
	}
	options := []protocol.LinkOption{{
		Spec:    "*",
		Label:   "Any version (*)",
		Meaning: "npm uses the library of this workspace whatever its version. You never edit the line when the library changes version. Best inside one workspace.",
	}}
	if v := params.Library.Version; v != "" {
		options = append(options, protocol.LinkOption{
			Spec:    "^" + v,
			Label:   "The current version (^" + v + ")",
			Meaning: "npm uses the workspace copy while its version fits the range, and the registry copy when it does not. Edit the line when the library moves to a new major version.",
		})
	}
	if len(options) == 1 {
		return options[0].Spec, nil
	}
	return options[0].Spec, options
}

// Unlink takes the project's dependency on the library out of its package.json
// and runs npm install. When the file did not list it, the result has no files.
func Unlink(workspaceRoot string, params protocol.UnlinkParams) (protocol.DependencyResult, error) {
	manifestPath := params.Project.Root + "/package.json"
	file := filepath.Join(workspaceRoot, filepath.FromSlash(manifestPath))
	raw, err := os.ReadFile(file)
	if err != nil {
		return protocol.DependencyResult{}, fmt.Errorf("%s: %w", manifestPath, err)
	}
	if len(params.Library.Exports) == 0 {
		return protocol.DependencyResult{}, errors.New(params.Library.Root + "/package.json has no \"name\"")
	}
	updated, from, err := RemoveDependency(raw, params.Library.Exports[0])
	if err != nil {
		return protocol.DependencyResult{}, fmt.Errorf("%s: %w", manifestPath, err)
	}
	if len(from) == 0 {
		return protocol.DependencyResult{}, nil
	}
	if err := os.WriteFile(file, updated, 0o644); err != nil {
		return protocol.DependencyResult{}, err
	}
	notes, err := dependencies.FoldedNpmRunner(workspaceRoot, append([]string{"npm", "install"}, quiet...))
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{Files: []string{manifestPath, "package-lock.json"}, Notes: notes}, nil
}
