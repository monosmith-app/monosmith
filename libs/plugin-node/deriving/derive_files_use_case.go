package deriving

import (
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/monosmith-app/monosmith/libs/plugin-node/siblingfiles"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Derive answers with the files the Node renderers derive for params.Projects.
// It reads files under workspaceRoot and writes none.
func Derive(workspaceRoot string, params protocol.DeriveParams) (protocol.DeriveResult, error) {
	byName := map[string]protocol.Project{}
	for _, p := range params.Projects {
		byName[p.Name] = p
	}
	// The shape of the sibling projects comes first: a file it derives is not
	// derived again by a renderer that knows one part of it.
	shared := siblingfiles.Derive(siblingfiles.Input{
		Root: workspaceRoot, Projects: params.Projects, Owned: params.Owned,
		References: func(p protocol.Project, content []byte) []any {
			return inFileOrder(dependencyReferences(workspaceRoot, p, byName, params), content)
		},
	})
	covered := map[string]bool{}
	files := shared.Files
	for _, f := range shared.Files {
		covered[f.Path] = true
	}
	for _, p := range params.Projects {
		if f, ok := tsconfigReferences(workspaceRoot, p, byName, params); ok && !covered[f.Path] {
			files = append(files, f)
		}
		if f, ok := jestDisplayName(workspaceRoot, p); ok && !covered[f.Path] {
			files = append(files, f)
		}
	}
	return protocol.DeriveResult{Files: files, Diverging: shared.Diverging}, nil
}

// readFile reads a workspace-relative file; ok is false when it is not there.
func readFile(workspaceRoot, rel string) ([]byte, bool) {
	b, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
	return b, err == nil
}

// relative is the path from one workspace-relative directory to another, with
// forward slashes.
func relative(from, to string) string {
	r, err := filepath.Rel(filepath.FromSlash(from), filepath.FromSlash(to))
	if err != nil {
		return path.Clean(to)
	}
	return filepath.ToSlash(r)
}

// owns reports whether the host's record owns the path.
func owns(params protocol.DeriveParams, rel string) bool {
	return slices.Contains(params.Owned, rel)
}
