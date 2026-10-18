package release

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestBareTagsGiveAnEmptyPrefixAndOtherProjectsKeepTheirs(t *testing.T) {
	java := &projectgraph.Project{Project: protocol.Project{Name: "opencraw-java"}, BareTags: true}
	other := &projectgraph.Project{Project: protocol.Project{Name: "opencraw"}}
	if got := TagPrefix(java); got != "" {
		t.Errorf("a bare-tag project's prefix = %q", got)
	}
	if got := TagPrefix(other); got != "opencraw@" {
		t.Errorf("another project's prefix = %q", got)
	}
}

func TestAPackageWithRegistryNonePublishesNowhere(t *testing.T) {
	p := &projectgraph.Project{
		Project: protocol.Project{Name: "opencraw-java", Kind: protocol.ProjectPackage, Registries: &protocol.Registries{NoPublish: true}},
		Plugin:  "no-such-plugin",
	}
	// The plugin is never asked: publish returns before calling it.
	res, err := publish(nil, p, Item{Next: "0.1.0"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Skipped != `published by its tag (registry = "none")` {
		t.Errorf("result %+v", res)
	}
}
