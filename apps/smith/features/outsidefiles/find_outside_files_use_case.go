package outsidefiles

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// msbuildExts are the MSBuild files of a project folder that Find reads.
var msbuildExts = map[string]bool{".csproj": true, ".fsproj": true, ".vbproj": true, ".props": true, ".targets": true}

// aboveFiles are the files MSBuild reads by itself from the folders above a
// project.
var aboveFiles = []string{"Directory.Build.props", "Directory.Build.targets", "Directory.Packages.props"}

// Find lists the files the project in sub uses outside sub, read from the
// whole source before the import takes the sub-folder only. It reads the
// MSBuild files and the package.json at the top of sub, and looks for
// Directory.Build.props, Directory.Build.targets and Directory.Packages.props
// in each folder from the parent of sub up to the top of the source. A path
// that leaves the source too is left out: nothing can bring it.
//
// Parameters: top is the top folder of the source on disk; sub is the
// project's sub-folder, relative to top, with forward slashes.
// Returns: the files, in the order of their manifests and then from the
// nearest folder above; empty when sub is the top or uses nothing outside.
func Find(top, sub string) []File {
	sub = path.Clean(strings.Trim(filepath.ToSlash(sub), "/"))
	if sub == "." || sub == "" {
		return nil
	}
	dir := filepath.Join(top, filepath.FromSlash(sub))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	var out []File
	seen := map[string]bool{}
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		var refs []Ref
		switch {
		case msbuildExts[strings.ToLower(filepath.Ext(name))]:
			refs = MSBuildRefs(string(raw))
		case name == "package.json":
			refs = PackageJSONRefs(raw)
		}
		for _, r := range refs {
			rel := path.Clean(path.Join(sub, strings.ReplaceAll(r.Value, `\`, "/")))
			if rel == ".." || strings.HasPrefix(rel, "../") || seen[rel] {
				continue
			}
			seen[rel] = true
			out = append(out, read(top, File{Ref: r.Value, Path: rel, Manifest: name, Kind: r.Kind}))
		}
	}
	if !hasMSBuild(names) {
		return out
	}
	for up := path.Dir(sub); ; up = path.Dir(up) {
		for _, name := range aboveFiles {
			rel := path.Join(up, name)
			if _, err := os.Stat(filepath.Join(top, filepath.FromSlash(rel))); err == nil {
				out = append(out, File{Path: rel, Kind: KindAbove})
			}
		}
		if up == "." {
			break
		}
	}
	return out
}

// read fills Content of a plain file, or Missing when the source does not
// hold it.
func read(top string, f File) File {
	raw, err := os.ReadFile(filepath.Join(top, filepath.FromSlash(f.Path)))
	switch {
	case err != nil:
		if _, statErr := os.Stat(filepath.Join(top, filepath.FromSlash(f.Path))); statErr != nil {
			f.Missing = true
		} else if f.Kind == KindFile {
			// A folder (an Include of a whole folder) cannot come along as a file.
			f.Kind = KindFolder
		}
	case f.Kind == KindFile:
		f.Content = raw
	}
	return f
}

// hasMSBuild reports whether a project folder holds an MSBuild project file.
func hasMSBuild(names []string) bool {
	for _, n := range names {
		switch strings.ToLower(filepath.Ext(n)) {
		case ".csproj", ".fsproj", ".vbproj":
			return true
		}
	}
	return false
}
