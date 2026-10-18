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

// teamRecipeAt is the team recipe in a version: kiwi, with a greeting that the
// plugin turns into a file it owns (kiwi.owned.txt) and a starter file
// (greeting.txt).
func teamRecipeAt(version, greeting string) string {
	return `name = "team"
version = "` + version + `"

[choices.language]
options = ["kiwi"]

[when.language.kiwi.use]
language = "kiwi"
toolchain = ["kiwi"]

[when.language.kiwi.settings.kiwi]
greeting = "` + greeting + `"
`
}

// TestRecipeDriftAndUpgrade follows a project through a recipe's new version
// with the real binary and an external plugin: doctor warns that the project
// has drifted and says what changed, upgrade --dry-run plans without writing,
// upgrade rewrites smith.toml and the file the plugin owns while leaving an
// edited starter file alone, a second upgrade changes nothing, and explain
// says which recipe line set each value.
func TestRecipeDriftAndUpgrade(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	plugin := filepath.Join(base, "kiwiplugin")
	if runtime.GOOS == "windows" {
		plugin += ".exe"
	}
	wd, _ := os.Getwd()
	run(t, filepath.Dir(wd), "go", "build", "-o", plugin, "./goldenpath/testdata/kiwiplugin")

	w.smith(true, "new", "acme", "--recipe", "recipes/team.recipe.toml")
	w.dir = filepath.Join(base, "acme")
	w.write("recipes/team.recipe.toml", teamRecipeAt("1.2.0", "hello from 1.2.0"))
	w.write(".smith/plugins/kiwi/plugin.json", `{"name": "kiwi", "version": "1.0.0", "command": [`+quote(filepath.ToSlash(plugin))+`]}`)
	w.write("smith.toml", w.read("smith.toml")+"\n[plugins.external.kiwi]\nversion = \"^1\"\n")
	w.write(".gitignore", w.read(".gitignore")+".kiwi-registry\n")

	w.smith(true, "add", "app", "basket", "--choice", "language=kiwi")
	if got := w.read("apps/basket/kiwi.owned.txt"); !strings.Contains(got, "greeting = hello from 1.2.0") {
		t.Fatalf("the plugin did not write the file it owns: %q", got)
	}
	if out := w.smith(true, "doctor"); !strings.Contains(out, "✔ basket follows its recipe team") {
		t.Fatalf("a project that follows its recipe passes:\n%s", out)
	}
	if out := w.smith(true, "upgrade"); !strings.Contains(out, "up to date") && strings.Contains(out, "upgraded basket") {
		t.Fatalf("a project that follows its recipe needs no upgrade:\n%s", out)
	}

	// The team edits the starter file; then the recipe moves to 1.3.0 and
	// changes the greeting.
	w.write("apps/basket/greeting.txt", "my own words\n")
	w.write("recipes/team.recipe.toml", teamRecipeAt("1.3.0", "hello from 1.3.0"))

	out := w.smith(true, "doctor")
	for _, want := range []string{
		"! basket follows its recipe team",
		"basket is on team 1.2.0; the recipe is 1.3.0, which changes the settings of the language plugin kiwi",
		`greeting: "hello from 1.2.0" → "hello from 1.3.0"`,
		"smith upgrade",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("doctor does not say %q:\n%s", want, out)
		}
	}

	// --dry-run plans, with the change to smith.toml, and writes nothing.
	before := w.read("smith.toml")
	out = w.smith(true, "upgrade", "--dry-run")
	for _, want := range []string{"would upgrade basket: recipe team 1.2.0 → 1.3.0", `-version = "1.2.0"`, `+version = "1.3.0"`, "would change: smith.toml"} {
		if !strings.Contains(out, want) {
			t.Errorf("the dry run does not say %q:\n%s", want, out)
		}
	}
	if w.read("smith.toml") != before || !strings.Contains(w.read("apps/basket/kiwi.owned.txt"), "1.2.0") {
		t.Fatal("--dry-run changed the workspace")
	}
	if out := w.smith(false, "upgrade", "--check"); !strings.Contains(out, "the workspace is not up to date") {
		t.Errorf("--check must fail while a project has drifted:\n%s", out)
	}

	// upgrade rewrites smith.toml and the owned file, and leaves the starter.
	out = w.smith(true, "upgrade")
	if !strings.Contains(out, "upgraded basket: recipe team 1.2.0 → 1.3.0") || !strings.Contains(out, "upgraded: apps/basket/kiwi.owned.txt") {
		t.Errorf("upgrade does not say what it did:\n%s", out)
	}
	cfg := w.read("smith.toml")
	for _, want := range []string{`version = "1.3.0"`, `greeting = "hello from 1.3.0"`} {
		if !strings.Contains(cfg, want) {
			t.Errorf("smith.toml lacks %q:\n%s", want, cfg)
		}
	}
	if strings.Contains(cfg, "1.2.0") || strings.Contains(cfg, "hello from 1.2.0") {
		t.Errorf("smith.toml still has the old recipe:\n%s", cfg)
	}
	if got := w.read("apps/basket/kiwi.owned.txt"); !strings.Contains(got, "greeting = hello from 1.3.0") {
		t.Errorf("the owned file was not rewritten: %q", got)
	}
	if got := w.read("apps/basket/greeting.txt"); got != "my own words\n" {
		t.Errorf("the starter file was touched: %q", got)
	}
	if out := w.smith(true, "run", "greet", "-p", "basket", "--no-cache"); !strings.Contains(out, "kiwi says hello from 1.3.0") {
		t.Errorf("the new setting did not reach the plugin's task:\n%s", out)
	}

	// A second upgrade changes nothing.
	snapshot := map[string]string{}
	for _, rel := range []string{"smith.toml", "apps/basket/kiwi.owned.txt", "apps/basket/greeting.txt", ".github/workflows/ci.yml"} {
		snapshot[rel] = w.read(rel)
	}
	if out := w.smith(true, "upgrade"); !strings.Contains(out, "up to date") {
		t.Errorf("a second upgrade must find nothing to do:\n%s", out)
	}
	for rel, text := range snapshot {
		if w.read(rel) != text {
			t.Errorf("the second upgrade changed %s", rel)
		}
	}
	if out := w.smith(true, "doctor"); !strings.Contains(out, "✔ basket follows its recipe team") {
		t.Errorf("after the upgrade the project follows its recipe:\n%s", out)
	}

	// explain names the recipe line behind each value.
	out = w.smith(true, "explain", "basket")
	for _, want := range []string{"recipe team 1.3.0", "settings.kiwi.greeting", `"hello from 1.3.0"`, "recipes/team.recipe.toml:", "(recipe team)"} {
		if !strings.Contains(out, want) {
			t.Errorf("explain does not say %q:\n%s", want, out)
		}
	}
	var explained struct {
		Recipe, Version string
		Values          []struct {
			Name   string
			Value  string
			Origin *struct {
				File string
				Line int
			}
		}
	}
	if err := json.Unmarshal([]byte(w.smith(true, "explain", "basket", "--recipe", "--json")), &explained); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, v := range explained.Values {
		if v.Name == "settings.kiwi.greeting" {
			found = v.Origin != nil && v.Origin.File == "recipes/team.recipe.toml" && v.Origin.Line > 0
		}
	}
	if explained.Recipe != "team" || explained.Version != "1.3.0" || !found {
		t.Errorf("explain --recipe --json: %+v", explained)
	}
}
