package scaffolding

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-applescript/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Name, Kind, Stack string
}

// Generate writes an AppleScript project: one main.applescript, which runs
// (an app) or offers a handler (a lib). There is no package: nothing
// publishes AppleScript.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind == protocol.ProjectPackage {
		return protocol.GenerateResult{}, errors.New("an AppleScript project is an app or a lib: there is no registry to publish to")
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	body := libTemplate
	note := "Load it from another script with: set lib to load script POSIX file \"<path>/" + params.Name + ".scpt\" (after smith build, on a Mac)."
	if params.Kind == protocol.ProjectApp {
		body = appTemplate
		note = "After smith build, on a Mac, run it with osascript " + protocol.ProjectBuildDir(params.Directory) + "/" + params.Name + ".scpt."
	}
	data := templateData{Name: params.Name, Kind: string(params.Kind), Stack: params.Stack}
	written, err := templating.Write(workspaceRoot, map[string]string{params.Directory + "/" + discovery.MainName: body}, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	return protocol.GenerateResult{Files: written, Notes: []string{note}}, nil
}
