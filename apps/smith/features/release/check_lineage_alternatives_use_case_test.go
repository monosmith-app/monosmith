package release

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/projectgraph"
)

func TestAlternativesTakeVTagsOnlyForTheOneReleasingProject(t *testing.T) {
	p := &projectgraph.Project{}
	p.Name = "font-color-contrast"
	p.Exports = []string{"font-color-contrast"}
	if got := alternatives(p, true); !slices.Contains(got, "v") {
		t.Errorf("the one package of a repository owns its v1.2.3 tags: %v", got)
	}
	if got := alternatives(p, false); slices.Contains(got, "v") {
		t.Errorf("with several packages a v tag belongs to none: %v", got)
	}
}
