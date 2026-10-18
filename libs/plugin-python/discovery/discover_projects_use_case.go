package discovery

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/pelletier/go-toml/v2"
)

// Discover describes the project behind each pyproject.toml. A root
// pyproject.toml configures tools and is not a project.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." {
			continue
		}
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		raw, err := os.ReadFile(filepath.Join(dir, "pyproject.toml"))
		if err != nil {
			return result, err
		}
		var py PyProject
		if err := toml.Unmarshal(raw, &py); err != nil {
			return result, fmt.Errorf("%s: %w", manifest, err)
		}
		if py.Project.Name == "" {
			continue // a tool-only pyproject (no [project] table) is not a project
		}
		imports, err := ScanImports(dir)
		if err != nil {
			return result, err
		}
		for _, req := range py.Project.Dependencies {
			if name := RequirementImportName(req); name != "" {
				imports = append(imports, name)
			}
		}
		sort.Strings(imports)
		importName := ImportName(py.Project.Name)
		kind := projectkind.Resolve(py.Tool.Smith.Kind, root, protocol.ProjectLib)
		result.Projects = append(result.Projects, protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "python",
			Kind:     kind,
			Stack:    py.Tool.Smith.Stack,
			Exports:  []string{importName},
			Imports:  dedupe(imports),
			Version:  py.Project.Version,
			Tasks:    Tasks(root, kind, importName, py.Tool.Mypy != nil),
		})
	}
	result.SetToolchain(func() string { return pythonToolchain() })
	return result, nil
}

func dedupe(sorted []string) []string {
	out := sorted[:0]
	for i, s := range sorted {
		if i == 0 || s != sorted[i-1] {
			out = append(out, s)
		}
	}
	return out
}
