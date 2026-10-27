package dependencies

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Rewrite changes declared version requirements in place, keeping the
// manifest's formatting, comments and every other key.
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
		out, ok := SetSpec(string(raw), c.Declaration.Section, c.Declaration.Name, c.Spec)
		if !ok {
			return protocol.DependencyResult{}, fmt.Errorf("%s no longer declares %s in [%s]", file, c.Declaration.Name, c.Declaration.Section)
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
