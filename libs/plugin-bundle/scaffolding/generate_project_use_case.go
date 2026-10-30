package scaffolding

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Name, Stack string
}

// Generate writes a bundle. It is an app: a lib or package has nothing to be
// assembled for.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind != protocol.ProjectApp {
		return protocol.GenerateResult{}, errors.New("a bundle is an app: it is assembled from other projects and released by its tag")
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	written, err := templating.Write(workspaceRoot, templates, templateData{Dir: params.Directory, Name: params.Name, Stack: params.Stack})
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	note := `List what it is built from in app.json "dependsOn", and give it a build task in smith.toml.`
	return protocol.GenerateResult{Files: written, Notes: []string{note}}, nil
}
