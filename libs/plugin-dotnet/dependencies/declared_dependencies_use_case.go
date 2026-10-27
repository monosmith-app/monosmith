package dependencies

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// workspaceFiles are the MSBuild files at the workspace root that declare
// packages for every project: central versions, and references shared by
// all projects.
var workspaceFiles = []string{"Directory.Packages.props", "Directory.Build.props"}

// ProjectFiles are the MSBuild project files of one project, workspace-
// relative: the .csproj or .fsproj in its directory, and its tests project,
// which lives beside it and is named by the project's test task.
func ProjectFiles(workspaceRoot string, p protocol.Project) []string {
	var files []string
	for _, pattern := range []string{"*.csproj", "*.fsproj"} {
		found, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), pattern))
		for _, f := range found {
			files = append(files, path.Join(p.Root, filepath.Base(f)))
		}
	}
	if t, ok := p.Tasks["test"]; ok && len(t.Command) > 2 {
		tests := path.Clean(path.Join(p.Root, t.Command[2]))
		if !contains(files, tests) {
			files = append(files, tests)
		}
	}
	sort.Strings(files)
	return files
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// Declared lists the package versions the workspace declares: the
// PackageReference of every project and its tests project, and the files at
// the workspace root (Directory.Packages.props with its central
// PackageVersion entries, Directory.Build.props). A reference without a
// version takes the central one and declares nothing of its own.
func Declared(workspaceRoot string, params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	var res protocol.DeclaredResult
	type site struct{ project, file string }
	var sites []site
	for _, f := range workspaceFiles {
		sites = append(sites, site{"", f})
	}
	for _, p := range params.Projects {
		for _, f := range ProjectFiles(workspaceRoot, p) {
			sites = append(sites, site{p.Name, f})
		}
	}
	for _, s := range sites {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(s.file)))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return res, err
		}
		packages, err := ParsePackages(raw)
		if err != nil {
			return res, fmt.Errorf("%s: %w", s.file, err)
		}
		for _, pkg := range packages {
			res.Declarations = append(res.Declarations, protocol.Declaration{
				Project: s.project, File: s.file, Name: pkg.Name, Spec: pkg.Version,
				Section: pkg.Element, Rewritable: Rewritable(pkg.Version),
			})
		}
	}
	sort.SliceStable(res.Declarations, func(i, j int) bool {
		a, b := res.Declarations[i], res.Declarations[j]
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Name < b.Name
	})
	return res, nil
}
