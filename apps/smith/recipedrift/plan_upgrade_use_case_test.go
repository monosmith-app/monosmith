package recipedrift

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/recipestack"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// planFor writes the smith.toml `smith add` would have written for orders
// (made from team 1.2.0 with ruff at line 100, and a hand-written task
// override), then plans the upgrade to the recipe given.
func planFor(t *testing.T, recipeText string) (Plan, *session.Session) {
	t.Helper()
	s, p, rec := sessionFor(t, recipeText)
	s.Graph = &projectgraph.Graph{Projects: []*projectgraph.Project{p}}
	toml := "[workspace]\nname = \"acme\"\nrecipe = \"team.recipe.toml\"\n\n" + recipestack.Render(recipestack.Declaration{
		Project:  "orders",
		Stack:    protocol.Stack{Name: "team-python", Description: "From the recipe team 1.2.0", Language: "python", Toolchain: []string{"ruff", "pytest"}},
		Settings: map[string]map[string]any{"ruff": {"line": int64(100)}},
		Recipe:   rec,
	}) + "\n[projects.orders.tasks.test]\ncache = false\n"
	if err := os.WriteFile(filepath.Join(s.Root, workspace.FileName), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanUpgrade(s, []Report{check(t, s, p, rec)})
	if err != nil {
		t.Fatal(err)
	}
	return plan, s
}

func TestPlanUpgradeUpdatesTheSettingsAndTheRecordedVersion(t *testing.T) {
	plan, _ := planFor(t, recipeAt("1.3.0", "", "ruff", 88))
	if !plan.Changed() || len(plan.Changes) != 1 || plan.Changes[0].Project != "orders" || plan.Changes[0].To != "1.3.0" {
		t.Fatalf("plan %+v", plan)
	}
	cfg, err := workspace.Parse([]byte(plan.After))
	if err != nil {
		t.Fatalf("%v\n%s", err, plan.After)
	}
	orders := cfg.Projects["orders"]
	if orders.Recipe.Version != "1.3.0" || orders.Settings["ruff"]["line"] != int64(88) {
		t.Errorf("record %+v settings %+v", orders.Recipe, orders.Settings)
	}
	if strings.Count(plan.After, "[stacks.team-python]") != 1 || !strings.Contains(plan.After, "From the recipe team 1.2.0") {
		t.Errorf("the stack is not changed when its plugins are not:\n%s", plan.After)
	}
	if orders.Tasks["test"].Cache == nil {
		t.Errorf("the hand-written task override was lost:\n%s", plan.After)
	}
	if strings.Contains(plan.Before, "line = 88") {
		t.Error("the plan must not touch the text it was made from")
	}
}

func TestPlanUpgradeRewritesTheStackWhenThePluginsChange(t *testing.T) {
	plan, _ := planFor(t, recipeAt("1.3.0", "", "flake8", 100))
	cfg, err := workspace.Parse([]byte(plan.After))
	if err != nil {
		t.Fatalf("%v\n%s", err, plan.After)
	}
	stack := cfg.Stacks["team-python"]
	if strings.Join(stack.Toolchain, ",") != "flake8,pytest" && strings.Join(stack.Toolchain, ",") != "pytest,flake8" {
		t.Errorf("stack %+v\n%s", stack, plan.After)
	}
	if !strings.Contains(stack.Description, "team 1.3.0") || strings.Contains(plan.After, "ruff") {
		t.Errorf("the old stack is still there:\n%s", plan.After)
	}
	if len(plan.Starters) != 1 || !strings.Contains(plan.Starters[0], "flake8") || !strings.Contains(plan.Starters[0], "starter files") {
		t.Errorf("an added plugin's starter files are not written, and the plan says so: %v", plan.Starters)
	}
}

func TestPlanUpgradeLeavesAStackAnotherProjectShares(t *testing.T) {
	s, p, rec := sessionFor(t, recipeAt("1.3.0", "", "flake8", 100))
	legacy := &projectgraph.Project{Project: protocol.Project{Name: "legacy", Kind: protocol.ProjectPackage, Stack: "team-python"}}
	s.Graph = &projectgraph.Graph{Projects: []*projectgraph.Project{p, legacy}}
	raw := "[workspace]\nname = \"acme\"\n"
	if err := os.WriteFile(filepath.Join(s.Root, workspace.FileName), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanUpgrade(s, []Report{check(t, s, p, rec)})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Changed() || plan.After != raw {
		t.Errorf("a shared stack must not be rewritten:\n%s", plan.After)
	}
	if len(plan.Skipped) != 1 || !strings.Contains(plan.Skipped[0], "orders is not upgraded") || !strings.Contains(plan.Skipped[0], "shared with legacy") {
		t.Errorf("skipped %v", plan.Skipped)
	}
}

func TestPlanUpgradeLeavesAStackAPluginProvides(t *testing.T) {
	s, p, rec := sessionFor(t, recipeAt("1.3.0", "", "flake8", 100))
	s.Graph = &projectgraph.Graph{Projects: []*projectgraph.Project{p}}
	s.Config.Stacks = nil
	s.Plugins.Clients[0].Info.Stacks = []protocol.Stack{{Name: "team-python", Language: "python", Toolchain: []string{"ruff", "pytest"}}}
	var err error
	if s.Stacks, err = pluginhost.NewCatalog(s.Plugins, s.Config); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.Root, workspace.FileName), []byte("[workspace]\nname = \"acme\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanUpgrade(s, []Report{check(t, s, p, rec)})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Changed() || len(plan.Skipped) != 1 || !strings.Contains(plan.Skipped[0], "provided by a plugin") {
		t.Errorf("plan %+v", plan)
	}
}

func TestPlanUpgradeReportsWhatCannotBeChecked(t *testing.T) {
	s, _, _ := sessionFor(t, recipeAt("1.2.0", "", "ruff", 100))
	s.Graph = &projectgraph.Graph{}
	if err := os.WriteFile(filepath.Join(s.Root, workspace.FileName), []byte("[workspace]\nname = \"acme\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanUpgrade(s, []Report{{Project: "orders", Recipe: "team", Problem: "the recipe could not be loaded"}, {Project: "ok", Recipe: "team", From: "1", To: "1"}})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Changed() || len(plan.Skipped) != 1 || !strings.Contains(plan.Skipped[0], "cannot be checked") {
		t.Errorf("plan %+v", plan)
	}
}
