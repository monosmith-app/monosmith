package scaffolding

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Name, Kind, Stack, Workspace string
}

// Generate writes a Lua project. An app also gets a script its rockspec
// installs; a lib is a module others require. smith does not publish to
// luarocks.org yet, so there is no package.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind == protocol.ProjectPackage {
		return protocol.GenerateResult{}, errors.New("a Lua project is an app or a lib: publishing to luarocks.org is not supported yet")
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	files := maps.Clone(templates)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(files, appTemplates)
	}
	data := templateData{Dir: params.Directory, Name: params.Name, Kind: string(params.Kind), Stack: params.Stack, Workspace: params.Workspace.Name}
	written, err := templating.Write(workspaceRoot, files, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	note := "Needs Lua with LuaRocks, luacheck and busted; fill in the source URL in the rockspec before publishing it."
	return protocol.GenerateResult{Files: written, Notes: []string{note}}, nil
}
