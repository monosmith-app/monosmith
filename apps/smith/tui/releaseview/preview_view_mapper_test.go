package releaseview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/release"
)

func TestAPreviewPlanSaysWhichMergeItShows(t *testing.T) {
	m := FromResult(Result{Plan: release.Plan{PreviewOf: "master", Unchanged: []string{"sdk"}}, Git: GitState{Known: true, Branch: "dev"}})
	if out := render(m); !strings.Contains(out, "Preview: what merging dev into master would release") {
		t.Fatalf("preview line missing:\n%s", out)
	}
	if out := render(FromResult(Result{Plan: release.Plan{Unchanged: []string{"sdk"}}})); strings.Contains(out, "Preview:") {
		t.Fatalf("an ordinary plan says it is a preview:\n%s", out)
	}
}
