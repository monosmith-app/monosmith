package recipestack

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func declare(t *testing.T, root string, d Declaration) {
	t.Helper()
	if _, err := Declare(root, d); err != nil {
		t.Fatal(err)
	}
}

func recordOf(version string) workspace.ProjectRecipe {
	return workspace.ProjectRecipe{Name: "team", Version: version, Source: "team.recipe.toml", Answers: map[string]string{"language": "python"}}
}

func TestRewriteBringsTheDeclarationsUpToDateAndKeepsTheRest(t *testing.T) {
	root := t.TempDir()
	base := "[workspace]\nname = \"acme\"\n\n# my comment\n[default_stacks]\nnode = \"node-jest\"\n"
	if err := os.WriteFile(filepath.Join(root, "smith.toml"), []byte(base), 0o644); err != nil {
		t.Fatal(err)
	}
	old := protocol.Stack{Name: "team-python", Description: "From the recipe team 1.2.0", Language: "python", Toolchain: []string{"ruff"}}
	declare(t, root, Declaration{Project: "billing", Stack: old, Settings: map[string]map[string]any{"ruff": {"line": 100}}, Recipe: recordOf("1.2.0")})
	declare(t, root, Declaration{Project: "invoices", Stack: old, Reuse: true, Settings: map[string]map[string]any{"ruff": {"line": 100}}, Recipe: recordOf("1.2.0")})
	// A hand-written override of a project the recipe made.
	raw, _ := os.ReadFile(filepath.Join(root, "smith.toml"))
	withTask := string(raw) + "\n[projects.billing.tasks.test]\ncache = false\n"

	now := protocol.Stack{Name: "team-python", Description: "From the recipe team 1.3.0", Language: "python", Toolchain: []string{"flake8"}}
	text, err := Rewrite(withTask, []Declaration{
		{Project: "billing", Stack: now, Settings: map[string]map[string]any{"flake8": {"max_line": 88}}, Recipe: recordOf("1.3.0")},
		{Project: "invoices", Stack: now, Reuse: true, Settings: map[string]map[string]any{"flake8": {"max_line": 88}}, Recipe: recordOf("1.3.0")},
	}, []string{"team-python"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(text, base) || !strings.Contains(text, "# my comment") {
		t.Errorf("the hand-written start was not kept:\n%s", text)
	}
	if !strings.Contains(text, "[projects.billing.tasks.test]\ncache = false") {
		t.Errorf("a hand-written override was lost:\n%s", text)
	}
	if n := strings.Count(text, "[stacks.team-python]"); n != 1 {
		t.Errorf("the stack is declared %d times:\n%s", n, text)
	}
	if strings.Contains(text, "ruff") || strings.Contains(text, "1.2.0") {
		t.Errorf("the old declaration is still there:\n%s", text)
	}
	if err := os.WriteFile(filepath.Join(root, "smith.toml"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := workspace.Load(root)
	if err != nil {
		t.Fatalf("%v\n%s", err, text)
	}
	if !reflect.DeepEqual(cfg.Stacks["team-python"].Toolchain, []string{"flake8"}) {
		t.Errorf("stack %+v", cfg.Stacks["team-python"])
	}
	for _, name := range []string{"billing", "invoices"} {
		p := cfg.Projects[name]
		if p.Recipe == nil || p.Recipe.Version != "1.3.0" || p.Settings["flake8"]["max_line"] != int64(88) || p.Settings["ruff"] != nil {
			t.Errorf("%s: %+v", name, p)
		}
	}
	if cfg.Projects["billing"].Tasks["test"].Cache == nil {
		t.Error("the task override no longer loads")
	}
}

func TestRewriteWithoutAReplacedStackLeavesTheStackAlone(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "smith.toml"), []byte("[workspace]\nname = \"acme\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	st := protocol.Stack{Name: "team-python", Language: "python", Toolchain: []string{"ruff"}}
	declare(t, root, Declaration{Project: "billing", Stack: st, Settings: map[string]map[string]any{"ruff": {"line": 100}}, Recipe: recordOf("1.2.0")})
	raw, _ := os.ReadFile(filepath.Join(root, "smith.toml"))
	text, err := Rewrite(string(raw), []Declaration{
		{Project: "billing", Stack: st, Reuse: true, Settings: map[string]map[string]any{"ruff": {"line": 88}}, Recipe: recordOf("1.3.0")},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(text, "[stacks.team-python]") != 1 || !strings.Contains(text, "line = 88") || strings.Contains(text, "line = 100") || !strings.Contains(text, `version = "1.3.0"`) {
		t.Errorf("settings and version only:\n%s", text)
	}
}
