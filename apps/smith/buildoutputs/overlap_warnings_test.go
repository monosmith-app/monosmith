package buildoutputs

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

func TestOverlapWarningsNameBothTasksAndTheFix(t *testing.T) {
	zip := projectWithOutputs("mvd", "apps/mvd", map[string][]string{"package-zip": {"../../dist/drop"}, "package-dmg": {"../../dist/drop"}})
	other := projectWithOutputs("tool", "apps/tool", map[string][]string{"package": {"../../dist/drop"}})
	projects := []*projectgraph.Project{zip, other}
	got := OverlapWarnings(projects)
	if len(got) == 0 {
		t.Fatal("no warning for tasks that write dist/drop")
	}
	for _, want := range []string{"mvd:package-dmg", "mvd:package-zip", "dist/drop", "shared_outputs"} {
		if !strings.Contains(strings.Join(got, "\n"), want) {
			t.Errorf("warnings miss %q:\n%s", want, strings.Join(got, "\n"))
		}
	}
	if got := OverlapWarnings(projects, "dist/drop"); len(got) != 0 {
		t.Errorf("a shared folder is not warned about: %v", got)
	}
}
