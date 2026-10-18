package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipedrift"
)

func TestRecipeFindingWarnsAboutDriftAndNamesTheRemedy(t *testing.T) {
	r := recipedrift.Report{Project: "orders", Recipe: "team", From: "1.2.0", To: "1.3.0", Changes: []string{"changes the settings of the lint plugin eslint (config: 1 → 2)"}}
	f := recipeFinding(r)
	if f.OK || !f.Warning || f.Check != "orders follows its recipe team" {
		t.Fatalf("drift is a warning: %+v", f)
	}
	if !strings.Contains(f.Detail, "orders is on team 1.2.0; the recipe is 1.3.0") || !strings.Contains(f.Remedy, "smith upgrade") {
		t.Errorf("detail %q remedy %q", f.Detail, f.Remedy)
	}
	if f.Fix != nil {
		t.Error("changing smith.toml is not a safe fix: the remedy is smith upgrade")
	}
}

func TestRecipeFindingPassesWhenTheProjectFollowsTheRecipe(t *testing.T) {
	f := recipeFinding(recipedrift.Report{Project: "orders", Recipe: "team", From: "1.3.0", To: "1.3.0"})
	if !f.OK || f.Warning {
		t.Errorf("%+v", f)
	}
}

func TestRecipeFindingReportsAProjectItCannotCheck(t *testing.T) {
	f := recipeFinding(recipedrift.Report{Project: "orders", Recipe: "team", Problem: "the recipe could not be loaded (no network)"})
	if f.OK || !f.Warning || !strings.Contains(f.Detail, "could not be loaded") || !strings.Contains(f.Remedy, "smith recipe check") {
		t.Errorf("an unreadable recipe is a warning, not a failure: %+v", f)
	}
}
