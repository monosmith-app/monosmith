package buildoutputs

import "testing"

import "github.com/monosmith-app/monosmith/apps/smith/projectgraph"

func TestInsideAnyMatchesThePathAndWhatIsUnderIt(t *testing.T) {
	allowed := []string{"apps/mvd/appwindow/web/"}
	for p, want := range map[string]bool{
		"apps/mvd/appwindow/web":      true,
		"apps/mvd/appwindow/web/a.js": true,
		"apps/mvd/appwindow/webextra": false,
		"apps/mvd/appwindow":          false,
	} {
		if got := InsideAny(p, allowed); got != want {
			t.Errorf("%s: got %v, want %v", p, got, want)
		}
	}
}

func TestASharedOutputFolderMayBeFilledByManyTasks(t *testing.T) {
	a := projectWithOutputs("a", "apps/a", map[string][]string{"package": {"../../dist/drop"}})
	b := projectWithOutputs("b", "apps/b", map[string][]string{"package": {"../../dist/drop"}})
	projects := []*projectgraph.Project{a, b}
	if got := SharedOutputs(projects); len(got) != 1 {
		t.Fatalf("without the setting it is reported: %v", got)
	}
	if got := SharedOutputs(projects, "dist/drop"); len(got) != 0 {
		t.Errorf("with the setting it is not: %v", got)
	}
}
