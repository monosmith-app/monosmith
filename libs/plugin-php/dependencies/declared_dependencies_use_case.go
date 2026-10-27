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

// Declared lists the Packagist packages each package's composer.json
// requires, in require and require-dev, and the workspace root's.
func Declared(workspaceRoot string, params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	var res protocol.DeclaredResult
	files := []struct{ project, file string }{{"", "composer.json"}}
	for _, p := range params.Projects {
		files = append(files, struct{ project, file string }{p.Name, path.Join(p.Root, "composer.json")})
	}
	for _, f := range files {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(f.file)))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return res, err
		}
		require, dev, err := ParseComposer(raw)
		if err != nil {
			return res, fmt.Errorf("%s: %w", f.file, err)
		}
		for section, packages := range map[string]map[string]string{SectionRequire: require, SectionDev: dev} {
			for name, constraint := range packages {
				res.Declarations = append(res.Declarations, protocol.Declaration{
					Project: f.project, File: f.file, Name: name, Spec: constraint, Section: section, Rewritable: Rewritable(constraint),
				})
			}
		}
	}
	sort.SliceStable(res.Declarations, func(i, j int) bool {
		a, b := res.Declarations[i], res.Declarations[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Section != b.Section {
			return a.Section < b.Section
		}
		return a.Name < b.Name
	})
	return res, nil
}
