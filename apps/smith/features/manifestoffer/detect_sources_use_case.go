package manifestoffer

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/adoptmove"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// maxDepth is how deep below the folder Sources and FindNested look.
const maxDepth = 4

// skipped are folders that hold no source of the project: installed
// dependencies, environments, build output and caches.
var skipped = map[string]bool{
	"node_modules": true, "venv": true, "__pycache__": true, "bin": true, "obj": true,
	"dist": true, "build": true, "target": true, "vendor": true,
}

// Sources lists, for each language plugin that names source patterns, the
// source files in dir and its folders (up to maxDepth deep, without hidden
// folders, installed dependencies and build output). A plugin that finds no
// file is left out. The language with the most files comes first.
//
// Parameters: dir is the folder; claims are the language plugins' claims.
// Returns: the finds, empty when no plugin recognises a file.
func Sources(dir string, claims []adoptmove.Claim) []Find {
	files := walk(dir, func(rel string, d fs.DirEntry) bool { return !d.IsDir() })
	var out []Find
	for _, c := range claims {
		if len(c.Sources) == 0 {
			continue
		}
		f := Find{Language: c.Language, Manifest: manifestWords(c.Manifests)}
		for _, rel := range files {
			if protocol.IsManifest(c.Sources, baseName(rel)) {
				f.Files = append(f.Files, rel)
			}
		}
		if len(f.Files) > 0 {
			out = append(out, f)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return len(out[i].Files) > len(out[j].Files) })
	return out
}

// FindNested lists the projects whose manifest is in a folder below dir (up
// to maxDepth deep, without hidden folders, installed dependencies and build
// output). A folder counts once, for the first claim whose manifest it holds.
//
// Parameters: dir is the source folder; claims are the language plugins'
// claims.
// Returns: the projects, by path.
func FindNested(dir string, claims []adoptmove.Claim) []Nested {
	seen := map[string]bool{}
	var out []Nested
	for _, rel := range walk(dir, func(rel string, d fs.DirEntry) bool { return !d.IsDir() && strings.Contains(rel, "/") }) {
		folder := filepath.ToSlash(filepath.Dir(filepath.FromSlash(rel)))
		if seen[folder] {
			continue
		}
		for _, c := range claims {
			if protocol.IsManifest(c.Manifests, baseName(rel)) {
				seen[folder] = true
				out = append(out, Nested{Path: folder, Manifest: baseName(rel), Language: c.Language})
				break
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// walk lists the entries below dir that keep accepts, relative and with
// forward slashes, in walk order.
func walk(dir string, keep func(rel string, d fs.DirEntry) bool) []string {
	var out []string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, rerr := filepath.Rel(dir, p)
		if rerr != nil || rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") || skipped[strings.ToLower(d.Name())] || strings.Count(rel, "/") >= maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		if keep(rel, d) {
			out = append(out, rel)
		}
		return nil
	})
	return out
}

// baseName is the base name of a relative path with forward slashes.
func baseName(rel string) string {
	return rel[strings.LastIndex(rel, "/")+1:]
}

// manifestWords names a plugin's first manifest as a person reads it.
func manifestWords(manifests []string) string {
	if len(manifests) == 0 {
		return "a manifest"
	}
	if m, ok := strings.CutPrefix(manifests[0], "*"); ok {
		return "a " + m + " file"
	}
	return manifests[0]
}
