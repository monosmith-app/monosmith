package dependencies

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// SectionRequire is where Go declares a module's requirements.
const SectionRequire = "require"

// Declared lists the modules each project's go.mod requires directly. The
// versions are written without the leading v (1.6.0), the way every
// ecosystem's declarations are compared. Indirect requirements are the go
// command's bookkeeping, not declarations, and are left out.
func Declared(workspaceRoot string, params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	var res protocol.DeclaredResult
	for _, p := range params.Projects {
		file := path.Join(p.Root, "go.mod")
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(file)))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return res, err
		}
		for _, r := range ParseRequirements(string(raw)) {
			if r.Indirect {
				continue
			}
			res.Declarations = append(res.Declarations, protocol.Declaration{
				Project: p.Name, File: file, Name: r.Path, Spec: strings.TrimPrefix(r.Version, "v"),
				Section: SectionRequire, Rewritable: Rewritable(r.Version),
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
