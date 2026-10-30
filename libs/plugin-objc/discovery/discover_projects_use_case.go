package discovery

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Discover describes the project behind each podspec. A project is named for
// its directory and exports its pod's name; the pods it depends on are its
// imports, of which the host keeps the ones another workspace project exports.
// A directory is one project: the first podspec in it counts, and a file with
// no pod name is not one. A pod is a lib unless its comment or directory says
// otherwise. A podspec at the workspace root is skipped, and so is one beside a
// Package.swift: that directory is a Swift package, which the swift plugin owns. Every project is
// tagged native: Xcode runs on macOS only.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	seen := map[string]bool{}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." || seen[root] {
			continue
		}
		if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(root), "Package.swift")); err == nil {
			continue // the swift plugin owns a directory with a Package.swift
		}
		src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(manifest)))
		if err != nil {
			return result, err
		}
		spec := ParsePodspec(string(src))
		if spec.Name == "" {
			continue
		}
		seen[root] = true
		result.Projects = append(result.Projects, protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "objc",
			Kind:     projectkind.Resolve(spec.Kind, root, protocol.ProjectLib),
			Exports:  []string{spec.Name},
			Imports:  spec.Dependencies,
			Version:  spec.Version,
			Tasks:    Tasks(path.Base(manifest)),
			Stack:    spec.Stack,
			Tags:     []string{NativeTag},
		})
	}
	return result, nil
}
