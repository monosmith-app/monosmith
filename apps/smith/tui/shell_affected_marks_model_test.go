package tui

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
)

func TestAllProjectsAffectedMeansNothingToCompareTo(t *testing.T) {
	all := []projectlist.Project{{Name: "a", Affected: true}, {Name: "b", Affected: true}}
	for _, p := range everythingIsNotNews(all) {
		if p.Affected {
			t.Fatalf("%s keeps a mark that says nothing", p.Name)
		}
	}
	if !all[0].Affected {
		t.Fatal("the input must stay as it was")
	}
	some := []projectlist.Project{{Name: "a", Affected: true}, {Name: "b"}}
	if !everythingIsNotNews(some)[0].Affected {
		t.Fatal("a real difference is kept")
	}
}
