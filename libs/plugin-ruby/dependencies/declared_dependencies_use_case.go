package dependencies

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Declared lists the gems with a version constraint that each gem's gemspec
// declares, and the workspace's Gemfile (the root's tools and test gems). A
// gem without a constraint has nothing to align and is left out.
func Declared(workspaceRoot string, params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	var res protocol.DeclaredResult
	add := func(project, file string, gems []GemRequirement) {
		for _, g := range gems {
			res.Declarations = append(res.Declarations, protocol.Declaration{
				Project: project, File: file, Name: g.Name, Spec: g.Constraint, Section: g.Section, Rewritable: Rewritable(g.Constraint),
			})
		}
	}
	read := func(file string) (string, bool, error) {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(file)))
		if errors.Is(err, fs.ErrNotExist) {
			return "", false, nil
		}
		return string(raw), err == nil, err
	}
	if src, ok, err := read("Gemfile"); err != nil {
		return res, err
	} else if ok {
		add("", "Gemfile", ParseGemfile(src))
	}
	for _, p := range params.Projects {
		found, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "*.gemspec"))
		for _, f := range found {
			file := path.Join(p.Root, filepath.Base(f))
			src, ok, err := read(file)
			if err != nil {
				return res, err
			}
			if ok {
				add(p.Name, file, ParseGemspec(src))
			}
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
