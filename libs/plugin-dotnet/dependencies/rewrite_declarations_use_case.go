package dependencies

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Rewrite changes package versions in place, in the MSBuild file that
// declares them (a project, Directory.Build.props or the central
// Directory.Packages.props), keeping the file's formatting and metadata.
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
		out, ok := SetVersion(string(raw), c.Declaration.Section, c.Declaration.Name, c.Spec)
		if !ok {
			return protocol.DependencyResult{}, fmt.Errorf("%s no longer declares %s", file, c.Declaration.Name)
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
