package dependencies

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Rewrite changes version constraints in place, in the gemspec or the
// Gemfile that declares them.
func Rewrite(workspaceRoot string, params protocol.RewriteParams) (protocol.DependencyResult, error) {
	var files []string
	seen := map[string]bool{}
	for _, c := range params.Changes {
		file := c.Declaration.File
		full := filepath.Join(workspaceRoot, filepath.FromSlash(file))
		raw, err := os.ReadFile(full)
		if err != nil {
			return protocol.DependencyResult{}, err
		}
		out, ok := SetConstraint(string(raw), c.Declaration.Name, Constraint(c.Declaration.Spec, c.Spec))
		if !ok {
			return protocol.DependencyResult{}, fmt.Errorf("%s no longer declares %s with a version", file, c.Declaration.Name)
		}
		if err := os.WriteFile(full, []byte(out), 0o644); err != nil {
			return protocol.DependencyResult{}, err
		}
		if !seen[file] {
			seen[file] = true
			files = append(files, file)
		}
	}
	return protocol.DependencyResult{Files: files}, nil
}
