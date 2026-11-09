package linking

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-go/dependencies"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// libraryModule reads the module path of the library: the identifier other
// modules require it by.
func libraryModule(workspaceRoot string, library protocol.Project) (string, error) {
	gomod := library.Root + "/go.mod"
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(gomod))); err != nil {
		return "", fmt.Errorf("%s is not there, so %s is not a Go module of its own and cannot be required: give it a go.mod, or declare nothing and import its package", gomod, library.Name)
	}
	if len(library.Exports) == 0 {
		return "", fmt.Errorf("%s has no module line", gomod)
	}
	return library.Exports[0], nil
}

// relativeFolder is the folder of target as seen from the folder from, with
// forward slashes, for a replace directive.
func relativeFolder(from, target string) string {
	rel, err := filepath.Rel(filepath.FromSlash(from), filepath.FromSlash(target))
	if err != nil {
		return "./" + target
	}
	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, ".") {
		rel = "./" + rel
	}
	return rel
}

// Link requires the library's module in the project's go.mod. Go finds the
// folder through go.work, which is made to list the library; with no go.work the
// file gets a replace directive that points at the folder. Nothing is fetched
// from a registry. With params.Plan nothing is written.
func Link(workspaceRoot string, params protocol.LinkParams) (protocol.LinkResult, error) {
	manifest := params.Project.Root + "/go.mod"
	file := filepath.Join(workspaceRoot, filepath.FromSlash(manifest))
	raw, err := os.ReadFile(file)
	if err != nil {
		return protocol.LinkResult{}, fmt.Errorf("%s: %w", manifest, err)
	}
	module, err := libraryModule(workspaceRoot, params.Library)
	if err != nil {
		return protocol.LinkResult{}, err
	}
	version := params.Spec
	if version == "" {
		version = placeholderVersion
	}
	res := protocol.LinkResult{Manifest: manifest, Line: "require " + module + " " + version}
	work, workErr := os.ReadFile(filepath.Join(workspaceRoot, "go.work"))
	if workErr != nil {
		res.Line += " and replace " + module + " => " + relativeFolder(params.Project.Root, params.Library.Root)
	}
	if params.Plan {
		return res, nil
	}
	text, changed := AddRequire(string(raw), module, version)
	if workErr != nil {
		var replaced bool
		text, replaced = AddReplace(text, module, relativeFolder(params.Project.Root, params.Library.Root))
		changed = changed || replaced
		res.Notes = append(res.Notes, "no go.work at the workspace root, so go.mod points at the folder with a replace directive")
	}
	needUse := workErr == nil && !WorkUses(string(work), params.Library.Root)
	if !changed && !needUse {
		res.Notes = []string{params.Project.Name + " already requires " + module}
		return res, nil
	}
	if changed {
		if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
			return protocol.LinkResult{}, err
		}
		res.Files = append(res.Files, manifest)
	}
	if needUse {
		if err := dependencies.GoRunner(workspaceRoot, []string{"go", "work", "use", "./" + path.Clean(params.Library.Root)}); err != nil {
			return protocol.LinkResult{}, err
		}
		res.Files = append(res.Files, "go.work")
	}
	return res, nil
}

// Unlink takes the module's require line, and the replace directive that went
// with it, out of the project's go.mod. go.work keeps listing the library,
// which other projects may use. When go.mod did not require it, the result has
// no files.
func Unlink(workspaceRoot string, params protocol.UnlinkParams) (protocol.DependencyResult, error) {
	manifest := params.Project.Root + "/go.mod"
	file := filepath.Join(workspaceRoot, filepath.FromSlash(manifest))
	raw, err := os.ReadFile(file)
	if err != nil {
		return protocol.DependencyResult{}, fmt.Errorf("%s: %w", manifest, err)
	}
	module, err := libraryModule(workspaceRoot, params.Library)
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	text, changed := RemoveRequire(string(raw), module)
	if !changed {
		return protocol.DependencyResult{}, nil
	}
	if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{Files: []string{manifest}, Notes: []string{"go.work still lists " + params.Library.Root + ", which other projects may use"}}, nil
}
