package convertplan

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

func TestPlanBlocksAFolderWithTwoLanguageManifests(t *testing.T) {
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{
		project("nx", "packages/nx", "lib", []string{"package.json", "Cargo.toml", "project.json"}, nil),
		project("ok", "packages/ok", "lib", []string{"package.json", "project.json"}, nil),
	}}
	plan := MakePlan(w, Input{DirName: "x", Read: func(string) string { return "" }})
	got := findings(plan, Blocker)
	if !plan.Blocked(false) || !strings.Contains(got, "packages/nx has package.json and Cargo.toml") || strings.Contains(got, "packages/ok") {
		t.Fatalf("blockers:\n%s", got)
	}
}
