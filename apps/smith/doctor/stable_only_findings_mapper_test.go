package doctor

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestStableOnlyIsReportedOnlyWithChannels(t *testing.T) {
	p := &projectgraph.Project{Project: protocol.Project{Name: "cli", Kind: protocol.ProjectApp}}
	p.Release = &workspace.ProjectRelease{PyPI: "acme", NuGet: "acme", StableOnly: []string{"pypi"}}
	projects := []*projectgraph.Project{p}
	if got := stableOnlyFindings(projects, false); len(got) != 0 {
		t.Fatalf("no channel, no pre-release: %+v", got)
	}
	got := stableOnlyFindings(projects, true)
	if len(got) != 2 || !got[0].Warning || got[0].Remedy == "" || !got[1].OK {
		t.Fatalf("want a warning for nuget and a note for pypi: %+v", got)
	}
}
