package scaffolding

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-vba/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Generate writes a VBA project: one main.bas, which has a Main procedure (an
// app) or a Greet function (a lib). There is no package: nothing publishes VBA.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind == protocol.ProjectPackage {
		return protocol.GenerateResult{}, errors.New("a VBA project is an app or a lib: there is no registry to publish to")
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return protocol.GenerateResult{}, err
	}
	body := source(moduleName(params.Name), params.Name, string(params.Kind), params.Stack)
	if err := os.WriteFile(filepath.Join(dir, discovery.MainName), []byte(body), 0o644); err != nil {
		return protocol.GenerateResult{}, err
	}
	note := "Import " + path.Join(params.Directory, discovery.MainName) + " into an Office document's VBA project (File > Import File in the VBA editor); export the modules back here to keep them in git."
	return protocol.GenerateResult{Files: []string{path.Join(params.Directory, discovery.MainName)}, Notes: []string{note}}, nil
}
