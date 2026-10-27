package dependencies

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Rewrite changes dependency versions in place, in the pom.xml or Gradle
// build script that declares them, keeping everything else as written.
func Rewrite(workspaceRoot string, params protocol.RewriteParams) (protocol.DependencyResult, error) {
	var files []string
	seen := map[string]bool{}
	for _, c := range params.Changes {
		d := c.Declaration
		full := filepath.Join(workspaceRoot, filepath.FromSlash(d.File))
		raw, err := os.ReadFile(full)
		if err != nil {
			return protocol.DependencyResult{}, err
		}
		var out string
		var ok bool
		if path.Base(d.File) == "pom.xml" {
			out, ok = SetPomVersion(string(raw), d.Section, d.Name, c.Spec)
		} else {
			out, ok = SetGradleVersion(string(raw), d.Name, d.Spec, c.Spec)
		}
		if !ok {
			return protocol.DependencyResult{}, fmt.Errorf("%s no longer declares %s", d.File, d.Name)
		}
		if err := os.WriteFile(full, []byte(out), 0o644); err != nil {
			return protocol.DependencyResult{}, err
		}
		if !seen[d.File] {
			seen[d.File] = true
			files = append(files, d.File)
		}
	}
	return protocol.DependencyResult{Files: files}, nil
}
