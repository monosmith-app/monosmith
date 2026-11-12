package scaffolding

import (
	"cmp"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Name, Target, Library, Kind, Stack string
}

// identifier is a C++ identifier part: lowercase letters, digits and
// underscores, starting with a letter (web-api → web_api).
func identifier(name string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(name) {
		switch {
		case unicode.IsLetter(c) || unicode.IsDigit(c):
			if b.Len() == 0 && unicode.IsDigit(c) {
				b.WriteByte('x')
			}
			b.WriteRune(c)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "_"):
			b.WriteByte('_')
		}
	}
	return strings.TrimSuffix(b.String(), "_")
}

// Target is a project's CMake target and namespace: the workspace's name
// first, so targets never clash when one project adds another (acme_core).
// workspace is the package prefix (protocol.WorkspaceInfo.PackagePrefix: the scope, else the name).
func Target(workspace, name string) string {
	if workspace == "" {
		return identifier(name)
	}
	return identifier(workspace) + "_" + identifier(name)
}

// Generate writes a CMake project: a library with the greeting slice, an
// app's executable, and a test CTest runs.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	target := cmp.Or(params.PackageName, Target(params.Workspace.PackagePrefix(), params.Name))
	data := templateData{Dir: params.Directory, Name: params.Name, Target: target, Library: target, Kind: string(params.Kind), Stack: params.Stack}
	templates := maps.Clone(sliceTemplates)
	maps.Copy(templates, listsTemplate)
	if params.Kind == protocol.ProjectApp {
		// The executable takes the project's name; its code is a library
		// the tests link too.
		data.Library = target + "_lib"
		maps.Copy(templates, appTemplates)
	}
	files, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	up := strings.Repeat("../", strings.Count(params.Directory, "/")+1)
	return protocol.GenerateResult{Files: files, Notes: []string{
		"Use it from another project with add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/" + up + params.Directory + " " + target + ") and target_link_libraries(<target> PRIVATE " + data.Library + ").",
	}}, nil
}
