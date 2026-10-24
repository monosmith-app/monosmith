package clicontract

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestWorkspaceSummaryAnswersWhatIsGoingOnHere(t *testing.T) {
	cfg := workspace.Config{
		Workspace:     workspace.WorkspaceSection{Name: "demo"},
		CI:            workspace.CISection{Providers: []string{"github", "azure"}},
		Release:       workspace.ReleaseSection{Branch: "main"},
		DefaultStacks: map[string]string{"node": "node-jest"},
	}
	graph := &projectgraph.Graph{Projects: []*projectgraph.Project{
		{Project: protocol.Project{Name: "web", Kind: "app", Stack: "node-jest"}, Plugin: "node"},
		{Project: protocol.Project{Name: "kit", Kind: "package"}, Plugin: "node"},
		{Project: protocol.Project{Name: "api", Kind: "app"}, Plugin: "go"},
	}}
	out := WorkspaceSummary("/work/demo", cfg, graph)
	for _, want := range []string{
		"workspace demo (/work/demo)",
		"projects  3 (2 apps, 1 package)",
		"languages 1 go, 2 node",
		"stacks    1 node-jest; default node-jest for node",
		"CI        github, azure",
		"releases  from main",
		"smith config list",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("summary lacks %q:\n%s", want, out)
		}
	}
}
