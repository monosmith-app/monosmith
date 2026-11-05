package projectgraph

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTheGitHubReleaseRegistryMarksAPackageToPack(t *testing.T) {
	cfg := workspace.Config{
		Release:  workspace.ReleaseSection{Registry: workspace.GitHubReleaseRegistry},
		Projects: map[string]workspace.ProjectOverride{"tagged": {Registry: workspace.NoRegistry}},
	}
	r := ProjectRegistries(cfg, "core")
	if r == nil || !r.PackToRelease || r.NoPublish || (protocol.Project{Kind: protocol.ProjectPackage, Registries: r}).Publishes() {
		t.Fatalf("core is packed and published nowhere: %+v", r)
	}
	if r := ProjectRegistries(cfg, "tagged"); r == nil || r.PackToRelease || !r.NoPublish {
		t.Fatalf("a project's own value wins: %+v", r)
	}
}
