package doctor

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/buildoutputs"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// buildWritesFinding fails when a build script writes directories its build
// task does not declare as outputs: a cache hit restores only the declared
// ones, so the real folder stays empty and whatever reads it (a dependent's
// types, the package's own scripts) breaks on a warm cache. The fix declares
// the outputs in smith.toml, which only changes what the cache restores.
func buildWritesFinding(s *session.Session) protocol.Finding {
	gaps := buildoutputs.FindGaps(s.Root, s.Graph.Projects)
	finding := protocol.Finding{Check: "every build declares the directories it writes as outputs", OK: len(gaps) == 0}
	if len(gaps) == 0 {
		return finding
	}
	var problems, remedies, projects []string
	for _, g := range gaps {
		problems = append(problems, g.Problem())
		remedies = append(remedies, g.Remedy())
		projects = append(projects, g.Project)
	}
	finding.Detail = Summary(problems, "; ", shownProblems)
	finding.Remedy = Summary(remedies, "\n", shownProblems)
	if len(remedies) > shownProblems {
		finding.Remedy += "\n`smith doctor --fix` declares them all"
	}
	if raw, err := os.ReadFile(filepath.Join(s.Root, workspace.FileName)); err == nil {
		finding.Fix = &protocol.Fix{
			Description: "declare the directories the builds write as their outputs in smith.toml",
			Safe:        true,
			AppliesTo:   projects,
			Files:       []protocol.FileEdit{{Path: workspace.FileName, Content: buildoutputs.DeclareOutputs(string(raw), gaps)}},
		}
	}
	return finding
}
