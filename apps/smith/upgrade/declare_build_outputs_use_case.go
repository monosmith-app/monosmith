package upgrade

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/buildoutputs"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// declareBuildOutputs fixes a workspace converted by an older smith, whose
// build scripts write to the project's own dist/ while the build task caches
// the workspace's dist/<project dir>: a cache hit then restored the wrong
// folder and left dist/ empty. Each project's build outputs in smith.toml
// become what its script writes. With check it changes nothing.
func declareBuildOutputs(s *session.Session, check bool, r *Result) error {
	gaps := buildoutputs.FindGaps(s.Root, s.Graph.Projects)
	if len(gaps) == 0 {
		return nil
	}
	names := make([]string, len(gaps))
	for i, g := range gaps {
		names[i] = g.Project + " (" + strings.Join(g.Outputs, ", ") + ")"
	}
	sort.Strings(names)
	r.Migrations = append(r.Migrations, "declared the build outputs the scripts write: "+strings.Join(names, ", "))
	addFile(r, workspace.FileName)
	if check {
		return nil
	}
	file := filepath.Join(s.Root, workspace.FileName)
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return os.WriteFile(file, []byte(buildoutputs.DeclareOutputs(string(raw), gaps)), 0o644)
}
