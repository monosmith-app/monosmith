package apisurface

import (
	"path"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// maxFollowDepth stops a chain of "export * from" that is longer than any real
// package, or a cycle the visited set did not catch.
const maxFollowDepth = 16

// List answers api/surface for a Node project: the names its entry points
// export at the revision in params (the working tree when the revision is
// empty). A project with nothing to list (private, no entry file, no
// package.json at the revision) is answered with Unreadable set, not an error.
func List(root string, params protocol.APISurfaceParams) (protocol.APISurfaceResult, error) {
	return list(&cachedSource{inner: NewSource(root, params.Ref), seen: map[string]cachedFile{}}, params.Project.Root), nil
}

// list reads the surface of the package in dir (workspace-relative) from src.
func list(src Source, dir string) protocol.APISurfaceResult {
	manifest, ok := src.Read(path.Join(dir, "package.json"))
	if !ok {
		return protocol.APISurfaceResult{Names: []string{}, Unreadable: "no package.json at this revision"}
	}
	points, private, err := EntryPoints(manifest)
	if err != nil {
		return protocol.APISurfaceResult{Names: []string{}, Unreadable: "package.json is not valid JSON"}
	}
	if private {
		return protocol.APISurfaceResult{Names: []string{}, Unreadable: "the package is private, so it has no public API"}
	}
	seen := map[string]bool{}
	var names []string
	found := false
	for _, point := range points {
		file, ok := entryFile(src, dir, point)
		if !ok {
			continue
		}
		found = true
		for _, name := range exportedFrom(src, file, map[string]bool{}, 0) {
			if point.Subpath != "." {
				name = point.Subpath + ":" + name
			}
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	if !found {
		return protocol.APISurfaceResult{Names: []string{}, Unreadable: "no entry point file found: neither the built file nor its source is in the repository"}
	}
	sort.Strings(names)
	if names == nil {
		names = []string{}
	}
	return protocol.APISurfaceResult{Names: names}
}

// entryFile picks the first candidate of an entry point that exists, or the
// source it was built from, and returns its workspace-relative path.
func entryFile(src Source, dir string, point EntryPoint) (string, bool) {
	for _, candidate := range point.Candidates {
		if _, ok := src.Read(path.Join(dir, candidate)); ok {
			return path.Join(dir, candidate), true
		}
	}
	for _, candidate := range point.Candidates {
		for _, variant := range sourceVariants(candidate) {
			if _, ok := src.Read(path.Join(dir, variant)); ok {
				return path.Join(dir, variant), true
			}
		}
	}
	return "", false
}

// exportedFrom lists the names a module exports, following "export * from" to
// the modules it re-exports. A re-export that cannot be followed (another
// package, a file that is missing) counts as one name, "* from <specifier>",
// so dropping it still shows as a removal.
func exportedFrom(src Source, file string, visited map[string]bool, depth int) []string {
	if visited[file] || depth > maxFollowDepth {
		return nil
	}
	visited[file] = true
	body, ok := src.Read(file)
	if !ok {
		return nil
	}
	exports := ParseExports(file, body)
	names := append([]string(nil), exports.Names...)
	for _, spec := range exports.Stars {
		next, ok := resolveModule(src, file, spec)
		if !ok {
			names = append(names, "* from "+spec)
			continue
		}
		names = append(names, exportedFrom(src, next, visited, depth+1)...)
	}
	return names
}
