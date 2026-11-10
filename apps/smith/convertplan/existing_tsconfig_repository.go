package convertplan

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

// existingTsconfig is the TypeScript config an @nx/js:tsc target without a
// tsConfig option compiles with: tsconfig.lib.json when the project has it,
// else tsconfig.json. When the project has neither, or the disk is not
// readable, it keeps Nx's own default, tsconfig.lib.json.
func existingTsconfig(w *nxworkspace.Workspace, p *nxworkspace.Project) string {
	if w != nil && w.Root != "" {
		for _, name := range []string{"tsconfig.lib.json", "tsconfig.json"} {
			if _, err := os.Stat(filepath.Join(w.Root, filepath.FromSlash(p.Root), name)); err == nil {
				return name
			}
		}
	}
	return "tsconfig.lib.json"
}
