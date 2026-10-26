package scaffolding

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Name, ID, Kind, Stack string
}

// pascal turns a kebab-case name into a .NET identifier part: web-api → WebApi.
func pascal(name string) string {
	var b strings.Builder
	upper := true
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			upper = true
			continue
		}
		if b.Len() == 0 && unicode.IsDigit(c) {
			b.WriteByte('X') // an identifier cannot start with a digit
		}
		if upper {
			c = unicode.ToUpper(c)
			upper = false
		}
		b.WriteRune(c)
	}
	return b.String()
}

// ProjectID is a project's assembly, namespace and package id:
// Workspace.Name (Acme.WebApi).
func ProjectID(workspace, name string) string {
	if workspace == "" {
		return pascal(name)
	}
	return pascal(workspace) + "." + pascal(name)
}

// Generate writes a project and its tests project, and the workspace's
// Directory.Build.props the first time.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	templates := maps.Clone(sliceTemplates)
	maps.Copy(templates, projectTemplate)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(templates, appTemplates)
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, directoryBuildProps)); errors.Is(err, fs.ErrNotExist) {
		maps.Copy(templates, rootTemplates)
	} else if err != nil {
		return protocol.GenerateResult{}, err
	}
	data := templateData{Dir: params.Directory, Name: params.Name, ID: ProjectID(params.Workspace.Name, params.Name), Kind: string(params.Kind), Stack: params.Stack}
	files, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	return protocol.GenerateResult{Files: files, Notes: []string{"Reference it from another project with <ProjectReference Include=\"<relative path>/" + data.ID + ".csproj\" />."}}, nil
}
