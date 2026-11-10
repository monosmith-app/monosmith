package discovery

import (
	"slices"

	"github.com/monosmith-app/monosmith/libs/plugin-go/dependencies"
)

// declaredRequirements adds the modules go.mod requires directly to the import
// paths scanned from the sources, without duplicates, so the host sees the
// dependency of a module on another module of the workspace even when no file
// imports it yet.
func declaredRequirements(imports []string, gomod string) []string {
	out := slices.Clone(imports)
	for _, r := range dependencies.ParseRequirements(gomod) {
		if !r.Indirect && !slices.Contains(out, r.Path) {
			out = append(out, r.Path)
		}
	}
	slices.Sort(out)
	return out
}
