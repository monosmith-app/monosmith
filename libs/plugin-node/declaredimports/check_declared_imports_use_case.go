package declaredimports

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
)

// Check lists the imports of the package at projectRoot (workspace-relative)
// that its package.json does not declare. A file that ships may import only
// what dependencies, peerDependencies (optional ones in peerDependenciesMeta
// too) or optionalDependencies declare; a
// test, e2e, tool or config file (DevFile) may also import devDependencies,
// the workspace root's declarations and what another workspace member declares. Node built-ins, the package's own
// name, and vscode in a VS Code extension are always fine. Versions are not
// compared, and a declaration nothing imports is not reported.
func Check(workspaceRoot, projectRoot string) ([]UndeclaredImport, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(projectRoot))
	pkg, err := readManifest(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil, err
	}
	// The root manifest is optional: a lone package has none above it.
	root, _ := readManifest(filepath.Join(workspaceRoot, "package.json"))
	shipped := names(pkg.Dependencies, pkg.PeerDependencies, pkg.OptionalDependencies)
	// An optional peer: the installing app provides it, or the package copes
	// without it (a guarded require).
	for name, meta := range pkg.PeerDependenciesMeta {
		if meta.Optional {
			shipped[name] = true
		}
	}
	dev := names(pkg.DevDependencies)
	workspaceDev := names(root.Dependencies, root.DevDependencies)
	siblings := declaredInWorkspace(workspaceRoot)
	byFile, err := discovery.ScanFileImports(dir)
	if err != nil {
		return nil, err
	}
	var out []UndeclaredImport
	for file, imported := range byFile {
		seen := map[string]bool{}
		for _, name := range imported {
			if seen[name] || builtins[name] || name == pkg.Name || (name == "vscode" && pkg.Engines.VSCode != "") || shipped[name] {
				continue
			}
			seen[name] = true
			if DevFile(file) && (dev[name] || workspaceDev[name] || siblings[name]) {
				continue
			}
			out = append(out, UndeclaredImport{File: file, Package: name, DevOnly: dev[name]})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Package < out[j].Package
	})
	return out, nil
}

func readManifest(file string) (discovery.PackageManifest, error) {
	var pkg discovery.PackageManifest
	raw, err := os.ReadFile(file)
	if err != nil {
		return pkg, err
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return pkg, fmt.Errorf("%s: %w", file, err)
	}
	return pkg, nil
}

func names(sections ...map[string]string) map[string]bool {
	out := map[string]bool{}
	for _, deps := range sections {
		for name := range deps {
			out[name] = true
		}
	}
	return out
}
