package release

import "testing"

func TestSiblingVersions(t *testing.T) {
	g := repo(t)
	pg := packages(t)
	pg.Get("core").Exports = []string{"@x/core"}
	pg.Get("sdk").Exports = []string{"@x/sdk"}
	commit(t, g, "packages/core/a", "feat(core): first feature")
	if err := g.CreateTag("core@0.4.0", "core 0.4.0"); err != nil {
		t.Fatal(err)
	}
	commit(t, g, "packages/sdk/a", "fix(sdk): only the sdk changes")

	plan, err := MakePlan(g, pg)
	if err != nil {
		t.Fatal(err)
	}
	if items := byProject(plan); len(items) != 1 || items["sdk"].Next == "" {
		t.Fatalf("only the sdk should release: %+v", plan.Items)
	}
	versions, err := SiblingVersions(g, pg, plan)
	if err != nil {
		t.Fatal(err)
	}
	if versions["@x/core"] != "0.4.0" {
		t.Fatalf("a sibling that is not releasing keeps its last released version, got %q", versions["@x/core"])
	}
	if versions["@x/sdk"] != byProject(plan)["sdk"].Next {
		t.Fatalf("a releasing package uses the version of this run, got %q", versions["@x/sdk"])
	}
}
