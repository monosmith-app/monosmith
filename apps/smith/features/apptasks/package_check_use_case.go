package apptasks

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node/declaredimports"
	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
	"github.com/spf13/cobra"
)

// PrebuiltEnv is the variable the package and package-check tasks of a
// function app set: the task graph built the app and its libraries first, so
// smith func package does not build them again.
const PrebuiltEnv = discovery.PrebuiltEnv

// checkDeclared checks that a function app and every workspace library its
// zip vendors declare in dependencies each package they import: the zip's
// production install installs only those, while the workspace's node_modules
// holds the others, so an undeclared import works here and breaks in the
// function host.
//
// Parameters: cmd prints the undeclared imports; f are the app's folders.
// Errors: a package is undeclared, or a manifest cannot be read.
func checkDeclared(cmd *cobra.Command, f Folders) error {
	root, err := workspace.FindRoot(f.Project)
	if err != nil {
		return err
	}
	folders := []string{f.Project}
	libraries, err := functionhost.WorkspaceLibraries(f.Project)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(libraries))
	for name := range libraries {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		folders = append(folders, libraries[name])
	}
	undeclared := 0
	for _, folder := range folders {
		rel, err := filepath.Rel(root, folder)
		if err != nil {
			return err
		}
		found, err := declaredimports.Check(root, filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		for _, u := range found {
			fmt.Fprintln(cmd.OutOrStdout(), u.String())
		}
		undeclared += len(found)
	}
	if undeclared > 0 {
		return fmt.Errorf("%d import(s) are not declared in dependencies, so the zip's production install leaves them out and the function host cannot load them: add each to the dependencies of the package.json named above", undeclared)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "the package of %s is laid out, and it and its %d librar(ies) declare every package they import\n", f.Name, len(names))
	return nil
}
