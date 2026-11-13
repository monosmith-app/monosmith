package scaffolding

import (
	"cmp"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectnaming"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Name, Module, Package, Kind, Stack string
}

// Generate writes a Swift package: an app is an executable target, a lib or
// a package a library target. A package differs from a lib only in being
// released for others to depend on by git tag, which the release does.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	files := maps.Clone(templates)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(files, appTemplates)
	} else {
		maps.Copy(files, libTemplates)
	}
	data := templateData{Dir: params.Directory, Name: params.Name, Module: projectnaming.Name(params.Workspace, projectnaming.Swift, params.Name), Package: cmp.Or(params.PackageName, projectnaming.Name(params.Workspace, projectnaming.Swift, params.Name)), Kind: string(params.Kind), Stack: params.Stack}
	written, err := templating.Write(workspaceRoot, files, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	note := "Needs the Swift toolchain (swift build, swift test)."
	if params.Kind == protocol.ProjectPackage {
		note += " Others depend on it by git tag: SwiftPM resolves semantic-version tags (1.2.3 or v1.2.3), not name@1.2.3, so set [projects." + params.Name + "] tag_prefix = \"v\" in smith.toml if this is the only package released from the repository."
	}
	return protocol.GenerateResult{Files: written, Notes: []string{note}}, nil
}
