package discovery

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Discover describes a project for each project file that is not a tests
// project. A tests project belongs to the project it references (its
// test task runs it); its references count as the project's own, for the
// graph. A project file at the workspace root is not a project of the
// monorepo.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	parsed := map[string]MSBuildProject{}
	byName := map[string]string{}
	for _, manifest := range manifests {
		src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(manifest)))
		if err != nil {
			return result, err
		}
		p, err := ParseProject(manifest, src)
		if err != nil {
			return result, fmt.Errorf("%s: %w", manifest, err)
		}
		parsed[manifest] = p
		byName[path.Clean(manifest)] = p.PackageID
	}
	// The tests project of each project, and the references it adds.
	tests := map[string]string{}
	testRefs := map[string][]string{}
	for manifest, p := range parsed {
		if !p.Test {
			continue
		}
		for _, ref := range p.References {
			target := path.Clean(path.Join(path.Dir(manifest), ref))
			if under, ok := parsed[target]; ok && !under.Test {
				tests[target] = manifest
				for _, r := range p.References {
					testRefs[target] = append(testRefs[target], path.Clean(path.Join(path.Dir(manifest), r)))
				}
			}
		}
	}
	for _, manifest := range manifests {
		p := parsed[manifest]
		root := path.Dir(manifest)
		if p.Test || root == "." {
			continue
		}
		fallback := protocol.ProjectLib
		if p.OutputType == "Exe" || p.OutputType == "WinExe" {
			fallback = protocol.ProjectApp
		}
		kind := projectkind.Resolve(p.Kind, root, fallback)
		imports := map[string]bool{}
		for _, ref := range p.References {
			if id, ok := byName[path.Clean(path.Join(root, ref))]; ok {
				imports[id] = true
			}
		}
		for _, ref := range testRefs[manifest] {
			if id, ok := byName[ref]; ok && id != p.PackageID {
				imports[id] = true
			}
		}
		var importList []string
		for id := range imports {
			importList = append(importList, id)
		}
		sort.Strings(importList)
		testsFile := ""
		if t, ok := tests[manifest]; ok {
			rel, _ := filepath.Rel(filepath.FromSlash(root), filepath.FromSlash(t))
			testsFile = filepath.ToSlash(rel)
		}
		result.Projects = append(result.Projects, protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "dotnet",
			Kind:     kind,
			Exports:  []string{p.PackageID},
			Imports:  importList,
			Version:  p.Version,
			Tasks:    tasksOf(workspaceRoot, root, path.Base(manifest), testsFile, kind),
			Stack:    p.Stack,
		})
	}
	result.SetToolchain(func() string { return toolrun.Toolchain("dotnet", "dotnet", "--version") })
	return result, nil
}
