package recipedrift

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func resolution(uses map[string][]string) recipe.Resolution {
	var res recipe.Resolution
	for key, plugins := range uses {
		res.Uses = append(res.Uses, recipe.Use{Key: key, Plugins: plugins})
	}
	return res
}

func TestStackChangesNamesThePluginsAndTheirLevel(t *testing.T) {
	old := protocol.Stack{Language: "python", Toolchain: []string{"pytest", "ruff"}, Structure: []string{"vertical-slices"}}
	now := protocol.Stack{Language: "python", Toolchain: []string{"pytest", "mypy"}, Deploy: []string{"docker"}}
	res := resolution(map[string][]string{"typecheck": {"mypy@^1"}, "deploy": {"docker"}})
	changes, added := stackChanges(old, now, res)
	want := []string{
		"adds the typecheck plugin mypy",
		"drops the toolchain plugin ruff",
		"drops the structure plugin vertical-slices",
		"adds the deploy plugin docker",
	}
	if !reflect.DeepEqual(changes, want) {
		t.Errorf("changes\n got %q\nwant %q", changes, want)
	}
	if !reflect.DeepEqual(added, []string{"mypy", "docker"}) {
		t.Errorf("added %v", added)
	}
}

func TestStackChangesIgnoresOrderAndReportsALanguageSwap(t *testing.T) {
	a := protocol.Stack{Language: "python", Toolchain: []string{"pytest", "ruff"}}
	b := protocol.Stack{Language: "python", Toolchain: []string{"ruff", "pytest"}}
	if changes, _ := stackChanges(a, b, recipe.Resolution{}); len(changes) != 0 {
		t.Errorf("order is not a change: %v", changes)
	}
	changes, _ := stackChanges(a, protocol.Stack{Language: "node", Scaffold: "typescript", Toolchain: a.Toolchain}, recipe.Resolution{})
	want := []string{"replaces the language plugin python with node", "changes the scaffold plugin from none to typescript"}
	if !reflect.DeepEqual(changes, want) {
		t.Errorf("got %q", changes)
	}
}

func TestSettingsChangesSpellsOutWhatChanged(t *testing.T) {
	old := map[string]map[string]any{
		"eslint": {"config": "@acme/eslint-config", "fix": true},
		"python": {"layout": "src"},
	}
	now := map[string]map[string]any{
		"eslint": {"config": "@acme/eslint-config/strict", "fix": true},
		"python": {"layout": "src"},
	}
	res := resolution(map[string][]string{"lint": {"eslint"}})
	st := protocol.Stack{Language: "python", Toolchain: []string{"eslint"}}
	got := settingsChanges(old, now, res, st, st)
	want := []string{`changes the settings of the lint plugin eslint (config: "@acme/eslint-config" → "@acme/eslint-config/strict")`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q", got)
	}
}

func TestSettingsChangesReadsNumbersAndNestedTablesAlike(t *testing.T) {
	// smith.toml gives int64, a freshly read recipe may give another type.
	old := map[string]map[string]any{"ruff": {"line": int64(100), "extend": map[string]any{"select": []any{"E", "F"}}}}
	same := map[string]map[string]any{"ruff": {"line": 100, "extend": map[string]any{"select": []string{"E", "F"}}}}
	if got := settingsChanges(old, same, recipe.Resolution{}, protocol.Stack{}, protocol.Stack{}); len(got) != 0 {
		t.Errorf("equal settings reported as changed: %v", got)
	}
	changed := map[string]map[string]any{"ruff": {"line": 88, "extend": map[string]any{"select": []string{"E"}}}}
	got := settingsChanges(old, changed, recipe.Resolution{}, protocol.Stack{}, protocol.Stack{})
	if len(got) != 1 || got[0] != `changes the settings of the plugin ruff (extend.select: ["E","F"] → ["E"], line: 100 → 88)` {
		t.Errorf("got %q", got)
	}
}

func TestSettingsChangesCoversAPluginAddedOrDroppedAndLongLists(t *testing.T) {
	old := map[string]map[string]any{"a": {"k1": 1, "k2": 1, "k3": 1, "k4": 1}}
	now := map[string]map[string]any{"a": {"k1": 2, "k2": 2, "k3": 2, "k4": 2}, "b": {"x": "y"}}
	got := settingsChanges(old, now, recipe.Resolution{}, protocol.Stack{}, protocol.Stack{})
	want := []string{
		"changes the settings of the plugin a (k1: 1 → 2, k2: 1 → 2, k3: 1 → 2, …)",
		`changes the settings of the plugin b (x: set to "y")`,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q", got)
	}
	got = settingsChanges(map[string]map[string]any{"a": {"k": 1}}, nil, recipe.Resolution{}, protocol.Stack{}, protocol.Stack{})
	if len(got) != 1 || got[0] != "changes the settings of the plugin a (k: removed)" {
		t.Errorf("dropped: %q", got)
	}
}

func TestAnswerChangesNamesTheNewChoicesOnly(t *testing.T) {
	got := answerChanges(map[string]string{"language": "python"}, map[string]string{"language": "python", "kind": "library"})
	if !reflect.DeepEqual(got, []string{`asks a new choice, kind, which takes its default "library"`}) {
		t.Errorf("got %q", got)
	}
	if got := answerChanges(map[string]string{"a": "1"}, map[string]string{"a": "1"}); len(got) != 0 {
		t.Errorf("unchanged answers: %v", got)
	}
}

func TestSummaryIsWrittenInTheRecipesTerms(t *testing.T) {
	r := Report{Project: "orders", Recipe: "team", From: "1.2.0", To: "1.3.0", Changes: []string{"changes the settings of the lint plugin eslint (config: 1 → 2)"}}
	if got, want := r.Summary(), "orders is on team 1.2.0; the recipe is 1.3.0, which changes the settings of the lint plugin eslint (config: 1 → 2)"; got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
	if !r.Drifted() {
		t.Error("a newer version is drift")
	}
	versionOnly := Report{Project: "orders", Recipe: "team", From: "1.2.0", To: "1.3.0"}
	if got := versionOnly.Summary(); got != "orders is on team 1.2.0; the recipe is 1.3.0, which changes nothing the project uses" {
		t.Errorf("version only: %s", got)
	}
	same := Report{Project: "orders", Recipe: "team", From: "1.3.0", To: "1.3.0", Changes: []string{"adds the lint plugin ruff"}}
	if got := same.Summary(); got != "orders differs from its recipe team 1.3.0: adds the lint plugin ruff" {
		t.Errorf("same version: %s", got)
	}
	ok := Report{Project: "orders", Recipe: "team", From: "1.3.0", To: "1.3.0"}
	if ok.Drifted() || ok.Summary() != "orders follows its recipe team 1.3.0" {
		t.Errorf("up to date: %v %s", ok.Drifted(), ok.Summary())
	}
	problem := Report{Project: "orders", Recipe: "team", Problem: "no network", From: "1", To: "2"}
	if problem.Drifted() || problem.Summary() != "orders cannot be checked against its recipe team: no network" {
		t.Errorf("problem: %v %s", problem.Drifted(), problem.Summary())
	}
}
