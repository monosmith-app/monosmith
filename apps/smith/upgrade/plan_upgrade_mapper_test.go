package upgrade

import "testing"

func TestPlanUpgradeGroupsMigrationsAndFiles(t *testing.T) {
	plan := PlanUpgrade(t.TempDir(), Result{Migrations: []string{"schema 1 to 2"}, Files: []string{".github/workflows/ci.yml"}})
	if len(plan.Groups) != 2 || plan.Groups[0].Heading != "smith.toml" || plan.Groups[1].Lines[0].Path != ".github/workflows/ci.yml" {
		t.Fatalf("%+v", plan)
	}
	if !PlanUpgrade(t.TempDir(), Result{}).Empty() {
		t.Fatal("an up-to-date workspace has an empty plan")
	}
}
