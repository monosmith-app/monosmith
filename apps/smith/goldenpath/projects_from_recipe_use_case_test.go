//go:build e2e

package goldenpath

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// recipeOfTheTeam is a two-language team recipe: the language decides the
// rest. Python is a built-in plugin; kiwi is an external one that records the
// settings the recipe gives it.
const recipeOfTheTeam = `name = "team"
version = "1.2.0"

[choices.language]
ask = "Language"
options = ["python", "kiwi"]

[choices.kind]
default = "library"
options = ["library", "service"]

[when.language.python.use]
language = "python"
structure = "vertical-slices"

[when.language.kiwi.use]
language = "kiwi"
toolchain = ["kiwi"]

[when.language.kiwi.settings.kiwi]
greeting = "hello from the team recipe"
`

// projectTasks lists a project's tasks as `smith projects --json` reports them.
func projectTasks(t *testing.T, w *workspace, name string) []string {
	t.Helper()
	var projects []struct {
		Name    string
		Targets []string
	}
	if err := json.Unmarshal([]byte(w.smith(true, "projects", "--json")), &projects); err != nil {
		t.Fatal(err)
	}
	for _, p := range projects {
		if p.Name == name {
			return p.Targets
		}
	}
	t.Fatalf("no project %s", name)
	return nil
}

func has(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

// TestProjectsFromRecipe creates a project of each language of a team recipe
// with `smith add --choice`, through the real binary and plugins: the recipe's
// plugins make the project, its settings reach the plugin, smith.toml records
// the stack and the recipe, and a stack or a bypass still works.
func TestProjectsFromRecipe(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	plugin := filepath.Join(base, "kiwiplugin")
	if runtime.GOOS == "windows" {
		plugin += ".exe"
	}
	wd, _ := os.Getwd()
	run(t, filepath.Dir(wd), "go", "build", "-o", plugin, "./goldenpath/testdata/kiwiplugin")

	// `smith new --recipe` names the recipe; one that is not in the new
	// workspace yet is said, not refused.
	out := w.smith(true, "new", "acme", "--recipe", "recipes/team.recipe.toml")
	if !strings.Contains(out, "not there yet") {
		t.Fatalf("new did not say the recipe is missing:\n%s", out)
	}
	w.dir = filepath.Join(base, "acme")
	if cfg := w.read("smith.toml"); !strings.Contains(cfg, `recipe = "recipes/team.recipe.toml"`) {
		t.Fatalf("smith.toml does not name the recipe:\n%s", cfg)
	}
	w.write("recipes/team.recipe.toml", recipeOfTheTeam)
	w.write(".smith/plugins/kiwi/plugin.json", `{"name": "kiwi", "version": "1.0.0", "command": [`+quote(filepath.ToSlash(plugin))+`]}`)
	w.write("smith.toml", w.read("smith.toml")+"\n[plugins.external.kiwi]\nversion = \"^1\"\n")
	w.write(".gitignore", w.read(".gitignore")+".kiwi-registry\n")

	// Without a terminal a choice with no default and no --choice fails,
	// naming it, before anything is written.
	before := w.read("smith.toml")
	out = w.smith(false, "add", "package", "billing")
	if !strings.Contains(out, `choice "language"`) || !strings.Contains(out, "--choice language=python|kiwi") {
		t.Fatalf("a missing answer must name the choice and the flag:\n%s", out)
	}
	if w.read("smith.toml") != before {
		t.Fatalf("a failed add changed smith.toml:\n%s", w.read("smith.toml"))
	}
	// --choice answers the recipe, and --stack bypasses it: not both.
	if out := w.smith(false, "add", "package", "billing", "--stack", "none", "--choice", "language=python"); !strings.Contains(out, "bypasses the recipe") {
		t.Fatalf("--stack with --choice:\n%s", out)
	}

	// Python: the built-in plugin makes the project, and the recipe's
	// structure plugin adds its check.
	out = w.smith(true, "add", "package", "billing", "--choice", "language=python")
	for _, want := range []string{"Recipe team 1.2.0 (language=python, kind=library) will use:", "structure", "vertical-slices", "Adding package billing (stack team-python-library)"} {
		if !strings.Contains(out, want) {
			t.Errorf("add printed no %q:\n%s", want, out)
		}
	}
	if _, err := os.Stat(filepath.Join(w.dir, "packages", "billing", "pyproject.toml")); err != nil {
		t.Fatalf("the python plugin did not make the project: %v", err)
	}
	if tasks := projectTasks(t, w, "billing"); !has(tasks, "slice-check") || !has(tasks, "test") {
		t.Fatalf("billing's tasks come from the recipe's plugins (language and structure): %v", tasks)
	}
	cfg := w.read("smith.toml")
	for _, want := range []string{
		"[stacks.team-python-library]", `language = "python"`, `structure = ["vertical-slices"]`,
		"[projects.billing.recipe]", `name = "team"`, `version = "1.2.0"`, `source = "recipes/team.recipe.toml"`,
		"[projects.billing.recipe.answers]", `language = "python"`, `kind = "library"`,
	} {
		if !strings.Contains(cfg, want) {
			t.Errorf("smith.toml lacks %q:\n%s", want, cfg)
		}
	}
	if !strings.Contains(w.read("packages/billing/pyproject.toml"), "team-python-library") {
		t.Errorf("the manifest does not record the stack:\n%s", w.read("packages/billing/pyproject.toml"))
	}

	// The same answers make the same stack: it is declared once.
	w.smith(true, "add", "package", "invoices", "--choice", "language=python", "--choice", "kind=library")
	if n := strings.Count(w.read("smith.toml"), "[stacks.team-python-library]"); n != 1 {
		t.Errorf("the stack is declared %d times", n)
	}

	// kiwi: an external plugin makes the project and is given the recipe's
	// settings, at generation and afterwards.
	out = w.smith(true, "add", "app", "basket", "--choice", "language=kiwi", "--choice", "kind=service")
	if !strings.Contains(out, "setting kiwi.greeting = hello from the team recipe") {
		t.Errorf("the plan did not show the recipe's setting:\n%s", out)
	}
	if got := strings.TrimSpace(w.read("apps/basket/greeting.txt")); got != "hello from the team recipe" {
		t.Fatalf("kiwi was not given the recipe's settings when generating: %q", got)
	}
	if tasks := projectTasks(t, w, "basket"); !has(tasks, "fmt") || !has(tasks, "greet") {
		t.Fatalf("basket's tasks come from the recipe's plugins (kiwi, with its settings): %v", tasks)
	}
	cfg = w.read("smith.toml")
	for _, want := range []string{"[stacks.team-kiwi-service]", `toolchain = ["kiwi"]`, "[projects.basket.settings.kiwi]", `greeting = "hello from the team recipe"`} {
		if !strings.Contains(cfg, want) {
			t.Errorf("smith.toml lacks %q:\n%s", want, cfg)
		}
	}
	if out := w.smith(true, "run", "greet", "-p", "basket", "--no-cache"); !strings.Contains(out, "kiwi says hello from the team recipe") {
		t.Errorf("the setting did not reach the plugin's task:\n%s", out)
	}
	if out := w.smith(true, "explain", "basket:greet"); !strings.Contains(out, "team-kiwi-service") {
		t.Errorf("explain does not name the recipe's stack:\n%s", out)
	}

	// --stack bypasses the recipe: an ordinary project, with nothing recorded.
	w.smith(true, "add", "package", "plain", "--lang", "python", "--stack", "none")
	if strings.Contains(w.read("smith.toml"), "[projects.plain") {
		t.Errorf("a project that bypassed the recipe was recorded:\n%s", w.read("smith.toml"))
	}
	if _, err := os.Stat(filepath.Join(w.dir, "packages", "plain", "pyproject.toml")); err != nil {
		t.Fatalf("the bypassing project is missing: %v", err)
	}
	// And the workspace is healthy with all of it.
	w.smith(true, "doctor")
}
