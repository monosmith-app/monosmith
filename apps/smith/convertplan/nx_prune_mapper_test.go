package convertplan

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

// azureHost is OpenCraw's apps/azure-host: an esbuild build into the app's
// own dist/, then Nx's prune (prune-lockfile and copy-workspace-modules).
func azureHost() *nxworkspace.Project {
	return &nxworkspace.Project{
		Name: "@opencraw/azure-host",
		Root: "apps/azure-host",
		Targets: map[string]nxworkspace.Target{
			"build":                  {Executor: "@nx/esbuild:esbuild", Options: map[string]json.RawMessage{"outputPath": json.RawMessage(`"apps/azure-host/dist"`)}},
			"prune-lockfile":         {Executor: "@nx/js:prune-lockfile"},
			"copy-workspace-modules": {Executor: "@nx/js:copy-workspace-modules"},
			"prune":                  {Executor: "nx:noop"},
		},
	}
}

func TestNxPruneBecomesTheNodePluginsPruneIntoTheBuildsDirectory(t *testing.T) {
	p := azureHost()
	plan := &Plan{}
	pp := &ProjectPlan{Name: "azure-host", Root: p.Root, Language: "node"}
	for _, name := range []string{"copy-workspace-modules", "prune", "prune-lockfile"} {
		if !IsNxPrune(name, p.Targets[name]) {
			t.Fatalf("%s is not taken for a prune step", name)
		}
		mapNxPrune(plan, p, pp, name)
	}
	task := plan.Override("azure-host").Tasks[pruneTask]
	if got := strings.Join(task.Command, " "); got != "{smith} npm-prune apps/azure-host --out apps/azure-host/dist" {
		t.Errorf("prune = %q", got)
	}
	if len(plan.Findings) != 3 || plan.Findings[0].Severity != Mapped || !strings.Contains(plan.Findings[0].Message, "into apps/azure-host/dist") {
		t.Errorf("findings %+v", plan.Findings)
	}
}

func TestNxPruneOfABuildIntoSmithsDirectoryNeedsNoOverride(t *testing.T) {
	p := azureHost()
	p.Targets["build"] = nxworkspace.Target{Executor: "@nx/esbuild:esbuild", Outputs: []string{"{workspaceRoot}/dist/{projectRoot}"}}
	plan := &Plan{}
	mapNxPrune(plan, p, &ProjectPlan{Name: "azure-host", Root: p.Root, Language: "node"}, "prune")
	if _, set := plan.Override("azure-host").Tasks[pruneTask]; set {
		t.Error("the plugin's own prune already writes into dist/apps/azure-host")
	}
}

func TestOnlyNxsPruneStepsArePrunes(t *testing.T) {
	if IsNxPrune("build", nxworkspace.Target{Executor: "nx:noop"}) || IsNxPrune("prune", nxworkspace.Target{Executor: "nx:run-commands"}) {
		t.Error("only the noop named prune, or a prune executor, is one")
	}
}
