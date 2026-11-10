package doctorchecks

import (
	"strconv"

	"github.com/monosmith-app/monosmith/apps/smith/buildoutputs"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// buildWritesFromGaps is the finding for the gaps between what builds write
// and what their tasks declare. config is the text of smith.toml, and
// canFix says it could be read: only then does the finding carry the fix.
func buildWritesFromGaps(gaps []buildoutputs.Gap, config string, canFix bool) protocol.Finding {
	finding := protocol.Finding{Check: "every build declares the directories it writes as outputs", OK: len(gaps) == 0, Area: AreaLint}
	if len(gaps) == 0 {
		return finding
	}
	var problems, remedies, projects []string
	for _, g := range gaps {
		problems = append(problems, g.Problem())
		remedies = append(remedies, g.Remedy())
		projects = append(projects, g.Project)
	}
	finding.Problem = problems[0]
	if len(problems) > 1 {
		finding.Problem = strconv.Itoa(len(problems)) + " builds write directories that are not declared as outputs"
	}
	finding.Detail = Summary(problems, "; ", ShownProblems)
	finding.Remedy = Summary(remedies, "\n", ShownProblems)
	if len(remedies) > ShownProblems {
		finding.Remedy += "\n`smith doctor --fix` declares them all"
	}
	if canFix {
		finding.Fix = &protocol.Fix{
			Description: "declare the directories the builds write as their outputs in smith.toml",
			Safe:        true,
			AppliesTo:   projects,
			Files:       []protocol.FileEdit{{Path: workspace.FileName, Content: buildoutputs.DeclareOutputs(config, gaps)}},
		}
	}
	return finding
}
