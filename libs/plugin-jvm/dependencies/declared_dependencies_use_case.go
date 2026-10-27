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

// BuildFiles are the build files of one module that can declare
// dependencies, in the order they are read.
var BuildFiles = []string{"pom.xml", "build.gradle.kts", "build.gradle"}

// Declared lists the dependency versions each module's pom.xml declares
// (directly, in dependencyManagement, or through a property of the pom) and
// each Gradle build script declares in the string notation. The versions
// are the build files', so Maven and Gradle keep resolving them themselves.
func Declared(workspaceRoot string, params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	var res protocol.DeclaredResult
	for _, p := range params.Projects {
		for _, name := range BuildFiles {
			file := path.Join(p.Root, name)
			raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(file)))
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			if err != nil {
				return res, err
			}
			var coordinates []Coordinate
			if name == "pom.xml" {
				if coordinates, err = ParsePom(raw); err != nil {
					return res, fmt.Errorf("%s: %w", file, err)
				}
			} else {
				coordinates = ParseGradle(raw)
			}
			for _, c := range coordinates {
				res.Declarations = append(res.Declarations, protocol.Declaration{
					Project: p.Name, File: file, Name: c.Name, Spec: c.Version, Section: c.Section, Rewritable: Rewritable(c.Version),
				})
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
