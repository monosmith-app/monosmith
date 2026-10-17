package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/libs/plugin-node/pruning"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Discover describes the project behind each package.json. The workspace's
// root manifest holds tooling only and is never a project.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." {
			continue
		}
		p, err := describe(workspaceRoot, root, manifest)
		if err != nil {
			return result, err
		}
		result.Projects = append(result.Projects, p)
	}
	return result, nil
}

func describe(workspaceRoot, root, manifest string) (protocol.Project, error) {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(manifest)))
	if err != nil {
		return protocol.Project{}, err
	}
	var pkg PackageManifest
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return protocol.Project{}, fmt.Errorf("%s: %w", manifest, err)
	}
	fallback := protocol.ProjectPackage
	if pkg.Private {
		fallback = protocol.ProjectLib
	}
	project := protocol.Project{
		Name:     path.Base(root),
		Root:     root,
		Language: "node",
		Kind:     projectkind.Resolve(pkg.Smith.Kind, root, fallback),
		Version:  pkg.Version,
		Stack:    pkg.Smith.Stack,
		Tasks:    map[string]protocol.Task{},
	}
	if pkg.Smith.Native {
		project.Tags = append(project.Tags, NativeTag)
	}
	if pkg.Name != "" {
		project.Exports = []string{pkg.Name}
	}
	imports, err := ScanImports(filepath.Join(workspaceRoot, filepath.FromSlash(root)))
	if err != nil {
		return project, err
	}
	project.Imports = mergeDeclared(imports, pkg)
	for script := range pkg.Scripts {
		project.Tasks[script] = TaskFor(script, root, pkg.Scripts)
	}
	// An app, or a VS Code extension, is packed into dist/drop: the
	// deliverable smith ci pack keeps (a .vsix for an extension).
	if _, own := pkg.Scripts["package"]; !own && (project.Kind == protocol.ProjectApp || pkg.Engines.VSCode != "") {
		project.Tasks["package"] = PackageTask(root, pkg.Scripts)
	}
	// An app that builds can be deployed from its build directory alone.
	if _, own := pkg.Scripts[pruning.Task]; !own && project.Kind == protocol.ProjectApp {
		if _, builds := pkg.Scripts["build"]; builds {
			project.Tasks[pruning.Task] = PruneTask(root)
		}
	}
	// What a published package imports is resolved where it is installed.
	if _, own := pkg.Scripts[DepsCheckTask]; !own && project.Kind == protocol.ProjectPackage {
		AddDepsCheck(&project)
	}
	return project, nil
}

// mergeDeclared adds the manifest's declared dependencies to the scanned
// imports, so a dependency used only by configuration still forms an edge.
func mergeDeclared(imports []string, pkg PackageManifest) []string {
	seen := map[string]bool{}
	for _, name := range imports {
		seen[name] = true
	}
	for _, deps := range []map[string]string{pkg.Dependencies, pkg.DevDependencies, pkg.PeerDependencies, pkg.OptionalDependencies} {
		for name := range deps {
			seen[name] = true
		}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
