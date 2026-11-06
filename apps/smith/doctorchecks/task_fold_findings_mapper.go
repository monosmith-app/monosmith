package doctorchecks

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/taskfolding"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TaskFoldFinding is TaskFoldChecks over the projects, the configuration and
// the text of smith.toml. The finding carries the fix only when the text can
// be rewritten without changing what any task resolves to.
func TaskFoldFinding(graph []*projectgraph.Project, cfg workspace.Config, src string) (protocol.Finding, bool) {
	var projects []taskfolding.Project
	for _, p := range graph {
		has := map[string]bool{}
		for name := range p.Tasks {
			// A task a project's own settings added is not the group's to change.
			has[name] = !strings.HasPrefix(p.Sources[name+".command"], "smith.toml [projects.")
		}
		projects = append(projects, taskfolding.Project{
			Name: p.Name, Language: p.Language, Kind: string(p.Kind),
			Has: has, Overrides: cfg.Projects[p.Name].Tasks,
		})
	}
	res := taskfolding.Fold(projects)
	if len(res.Folded) == 0 {
		return protocol.Finding{}, false
	}
	var parts []string
	var applies []string
	for _, f := range res.Folded {
		parts = append(parts, fmt.Sprintf("%s (%d projects, %s)", f.Task, len(f.Projects), strings.Join(f.Select, " and ")))
		applies = append(applies, f.Projects...)
	}
	sort.Strings(applies)
	finding := protocol.Finding{
		Check:   "task settings shared by projects are written once",
		Warning: true,
		Area:    AreaConfig,
		Problem: "projects repeat the same overrides in smith.toml: " + leadProblem(parts),
		Detail:  "projects repeat the same overrides in smith.toml: " + strings.Join(parts, "; "),
		Remedy:  "fold them into a [[groups]] entry (see docs/reference/configuration.md, Task groups); smith doctor --fix writes it and keeps your comments",
	}
	if text, ok := taskfolding.Rewrite(src, projects, res); ok {
		finding.Fix = &protocol.Fix{
			Description: "write the repeated task settings once, as groups",
			Safe:        true,
			Files:       []protocol.FileEdit{{Path: workspace.FileName, Content: text}},
			AppliesTo:   dedupe(applies),
		}
	}
	return finding, true
}

func dedupe(sorted []string) []string {
	var out []string
	for i, s := range sorted {
		if i == 0 || s != sorted[i-1] {
			out = append(out, s)
		}
	}
	return out
}
