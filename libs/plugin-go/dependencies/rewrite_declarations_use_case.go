package dependencies

import (
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Rewrite moves modules to the versions asked, one project at a time, with
// `go get module@version`: go.mod and go.sum change together, so the
// project still builds. It needs the module proxy (or the module cache) for
// the go.sum hashes.
func Rewrite(workspaceRoot string, params protocol.RewriteParams) (protocol.DependencyResult, error) {
	byFile := map[string][]string{}
	var order []string
	for _, c := range params.Changes {
		file := c.Declaration.File
		if _, ok := byFile[file]; !ok {
			order = append(order, file)
		}
		byFile[file] = append(byFile[file], c.Declaration.Name+"@v"+c.Spec)
	}
	var files []string
	for _, file := range order {
		root := path.Dir(file)
		argv := append([]string{"go", "get"}, byFile[file]...)
		if err := GoRunner(filepath.Join(workspaceRoot, filepath.FromSlash(root)), argv); err != nil {
			return protocol.DependencyResult{Files: files}, err
		}
		files = append(files, file, path.Join(root, "go.sum"))
	}
	return protocol.DependencyResult{Files: files}, nil
}
