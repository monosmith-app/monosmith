package apisurface

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/gitrevision"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// List answers api/surface for a Go project: the exported names of its module
// at the revision in params (the working tree when the revision is empty). A
// project with no go.mod at the revision is answered with Unreadable set, not
// an error.
//
// Parameters: root is the workspace root; params name the project and revision.
// Returns: the result, and no error: a failure to read is Unreadable.
func List(root string, params protocol.APISurfaceParams) (protocol.APISurfaceResult, error) {
	dir := params.Project.Root
	var files map[string]string
	if params.Ref == "" {
		files = readWorktree(root, dir)
	} else {
		reader := gitrevision.NewReader(root, params.Ref)
		defer func() { _ = reader.Close() }()
		files = readRevision(reader, dir)
	}
	if _, ok := files["go.mod"]; !ok {
		return protocol.APISurfaceResult{Names: []string{}, Unreadable: "no go.mod at this revision"}, nil
	}
	return protocol.APISurfaceResult{Names: ExportedNames(inThisModule(files))}, nil
}

// readWorktree reads the go.mod and the .go files below dir on disk, keyed by
// the path relative to dir.
func readWorktree(root, dir string) map[string]string {
	base := filepath.Join(root, filepath.FromSlash(dir))
	files := map[string]string{}
	_ = filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && (d.Name() == "node_modules" || d.Name() == ".git") {
			return filepath.SkipDir
		}
		if d.IsDir() || (!strings.HasSuffix(p, ".go") && filepath.Base(p) != "go.mod") {
			return nil
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(base, p)
		files[filepath.ToSlash(rel)] = string(body)
		return nil
	})
	return files
}

// readRevision reads the go.mod and the .go files below dir at the revision.
func readRevision(reader *gitrevision.Reader, dir string) map[string]string {
	files := map[string]string{}
	for _, p := range reader.List(dir) {
		if !strings.HasSuffix(p, ".go") && path.Base(p) != "go.mod" {
			continue
		}
		body, ok := reader.Read(p)
		if !ok {
			continue
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, strings.TrimPrefix(dir, "./")), "/")
		files[rel] = body
	}
	return files
}

// inThisModule drops the files of nested modules: a folder with its own go.mod
// is another module's API.
func inThisModule(files map[string]string) map[string]string {
	nested := map[string]bool{}
	for p := range files {
		if path.Base(p) == "go.mod" && path.Dir(p) != "." {
			nested[path.Dir(p)+"/"] = true
		}
	}
	out := map[string]string{}
	for p, body := range files {
		inNested := false
		for n := range nested {
			if strings.HasPrefix(p, n) {
				inNested = true
				break
			}
		}
		if !inNested {
			out[p] = body
		}
	}
	return out
}
