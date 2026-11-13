package functionhost

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// linkedFolder is the folder that npm links as node_modules/<name> for a
// library of the workspace, found from a folder up to the root. It is empty
// when the package is not a link.
func linkedFolder(from, name string) string {
	for dir := from; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "node_modules", filepath.FromSlash(name))
		if info, err := os.Lstat(candidate); err == nil && info.Mode()&os.ModeSymlink != 0 {
			if real, err := filepath.EvalSymlinks(candidate); err == nil {
				return real
			}
		}
		if filepath.Dir(dir) == dir {
			return ""
		}
	}
}

// workspaceRootOf is the folder with node_modules above a library: the root of
// the workspace.
func workspaceRootOf(folder string) string {
	for dir := filepath.Dir(folder); ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "node_modules")); err == nil {
			return dir
		}
		if filepath.Dir(dir) == dir {
			return filepath.Dir(folder)
		}
	}
}

// workspaceLibraries finds the libraries of the workspace that dependencies
// name, and the ones those libraries need too. found maps name to folder.
func workspaceLibraries(projectDir string, dependencies map[string]any, found map[string]string) error {
	names := make([]string, 0, len(dependencies))
	for name := range dependencies {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		if _, seen := found[name]; seen {
			continue
		}
		folder := linkedFolder(projectDir, name)
		if folder == "" {
			continue
		}
		found[name] = folder
		manifest, err := readManifest(filepath.Join(folder, "package.json"))
		if err != nil {
			return err
		}
		if err := workspaceLibraries(projectDir, stringMap(manifest["dependencies"]), found); err != nil {
			return err
		}
	}
	return nil
}

// vendorLibrary builds a library of the workspace, packs what it ships into
// vendorDir and returns the name of the tarball. A tarball is always installed
// as a copy, where a folder would be installed as a link that a zip cannot hold.
func vendorLibrary(tools Tools, name, folder, vendorDir string) (string, error) {
	root := workspaceRootOf(folder)
	pkg, err := readManifest(filepath.Join(folder, "package.json"))
	if err != nil {
		return "", err
	}
	if scripts := stringMap(pkg["scripts"]); scripts["build"] != nil {
		if err := tools.Run(folder, []string{"npm", "run", "build"}); err != nil {
			return "", err
		}
	}
	entry := "index.js"
	if s, ok := pkg["main"].(string); ok {
		entry = s
	} else if s, ok := pkg["exports"].(string); ok {
		entry = s
	}
	flat := strings.ReplaceAll(strings.TrimPrefix(name, "@"), "/", "-")
	target := filepath.Join(vendorDir, flat)
	if err := os.RemoveAll(target); err != nil {
		return "", err
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return "", err
	}
	manifest := map[string]any{}
	for k, v := range pkg {
		manifest[k] = v
	}
	for _, k := range []string{"scripts", "devDependencies", "files", "types"} {
		delete(manifest, k)
	}
	rel, err := filepath.Rel(root, folder)
	if err != nil {
		return "", err
	}
	builtHere := filepath.Join(root, "dist", rel)
	if err := placeLibraryCode(name, entry, folder, builtHere, target, manifest); err != nil {
		return "", err
	}
	if err := writeManifest(filepath.Join(target, "package.json"), manifest); err != nil {
		return "", err
	}
	packed, err := tools.Output(vendorDir, []string{"npm", "pack", "./" + flat, "--silent", "--ignore-scripts"})
	if err != nil {
		return "", err
	}
	if err := os.RemoveAll(target); err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(packed), "\n")
	return strings.TrimSpace(lines[len(lines)-1]), nil
}

// placeLibraryCode copies the code a library ships into target and points the
// manifest's main at it. The library is built into the workspace's dist/
// (builtHere), ships from its own folder, or its main points into dist/.
func placeLibraryCode(name, entry, folder, builtHere, target string, manifest map[string]any) error {
	isTS := strings.HasSuffix(entry, ".ts")
	source := strings.TrimSuffix(strings.TrimPrefix(entry, "./"), ".ts") + ".js"
	if isTS {
		for _, file := range []string{source, strings.TrimPrefix(source, "src/")} {
			if _, err := os.Stat(filepath.Join(builtHere, filepath.FromSlash(file))); err == nil {
				manifest["main"] = "./dist/" + file
				delete(manifest, "exports")
				return copyTree(builtHere, filepath.Join(target, "dist"))
			}
		}
		return fmt.Errorf("%s is shipped as TypeScript (%s) and has no build in %s, so a function host cannot run it: give the library a build script that writes there", name, entry, builtHere)
	}
	out := filepath.Dir(filepath.Join(folder, filepath.FromSlash(entry)))
	if rel, err := filepath.Rel(folder, out); err == nil && strings.HasPrefix(rel, "..") {
		manifest["main"] = "./dist/" + filepath.Base(entry)
		delete(manifest, "exports")
		return copyTree(out, filepath.Join(target, "dist"))
	}
	return copyTree(folder, target)
}
