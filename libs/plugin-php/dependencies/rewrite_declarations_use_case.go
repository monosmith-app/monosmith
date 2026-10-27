package dependencies

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/libs/plugin-php/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Rewrite changes version constraints in place in the composer.json that
// declares them, keeping its formatting. When the workspace has a
// composer.lock it then moves the changed packages in the lock too
// (`composer update --no-install`), because composer install would
// otherwise keep installing the locked versions; the next install brings
// vendor/ to them.
func Rewrite(workspaceRoot string, params protocol.RewriteParams) (protocol.DependencyResult, error) {
	var files []string
	seen := map[string]bool{}
	names := map[string]bool{}
	for _, c := range params.Changes {
		file := c.Declaration.File
		full := filepath.Join(workspaceRoot, filepath.FromSlash(file))
		raw, err := os.ReadFile(full)
		if err != nil {
			return protocol.DependencyResult{}, err
		}
		out, ok := SetConstraint(string(raw), c.Declaration.Section, c.Declaration.Name, c.Spec)
		if !ok {
			return protocol.DependencyResult{}, fmt.Errorf("%s no longer declares %s in %s", file, c.Declaration.Name, c.Declaration.Section)
		}
		if err := os.WriteFile(full, []byte(out), 0o644); err != nil {
			return protocol.DependencyResult{}, err
		}
		names[c.Declaration.Name] = true
		if !seen[file] {
			seen[file] = true
			files = append(files, file)
		}
	}
	res := protocol.DependencyResult{Files: files}
	if _, err := os.Stat(filepath.Join(workspaceRoot, "composer.lock")); err != nil || len(names) == 0 {
		return res, nil
	}
	argv := []string{"composer", "update", "--no-interaction", "--no-progress", "--no-install", "--with-dependencies"}
	var sorted []string
	for n := range names {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)
	if err := scaffolding.Runner(workspaceRoot, append(argv, sorted...)); err != nil {
		return res, err
	}
	res.Files = append(res.Files, "composer.lock")
	return res, nil
}
