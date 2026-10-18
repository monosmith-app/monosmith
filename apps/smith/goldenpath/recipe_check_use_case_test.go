//go:build e2e

package goldenpath

import (
	"encoding/json"
	"strings"
	"testing"
)

const companyRecipe = `name = "company"
version = "3.0.0"

[settings.eslint]
config = "@monosmith/eslint-config"

[settings.dotnet]
target_framework = "net8.0"
`

// teamRecipe is the design's example: one recipe for a team that works in
// Python, C# and Node, where the language decides the rest.
const teamRecipe = `name = "team"
version = "1.2.0"
extends = "company/base.recipe.toml"

[choices.language]
ask = "Language"
options = ["python", "csharp", "node"]

[choices.kind]
ask = "What are you creating?"
options = ["library", "service"]
default = "library"

[when.language.python.use]
language = "python"
structure = "vertical-slices"

[when.language.csharp.use]
language = "dotnet"
structure = "vertical-slices"

[when.language.csharp.settings.dotnet]
target_framework = "net10.0"

[when.language.csharp.when.kind.service.use]
deploy = "docker"

[when.language.node.use]
language = "node"
lint = "eslint"
test = "vitest"
`

// TestRecipeCheck resolves a team recipe for each language its choice offers,
// through the real binary and the real built-in plugins: the plugins and
// settings it uses, where each came from, and what it refuses.
func TestRecipeCheck(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.write("company/base.recipe.toml", companyRecipe)
	w.write("team.recipe.toml", teamRecipe)

	for _, tc := range []struct {
		args []string
		want []string
		not  []string
	}{
		{[]string{"--choice", "language=python"},
			[]string{"Recipe team 1.2.0", "chain: company 3.0.0", "language     python", "structure    vertical-slices", "eslint.config", "base.recipe.toml:5 (recipe company)", "The recipe is valid."},
			[]string{"deploy"}},
		{[]string{"--choice", "language=csharp"},
			[]string{"language     dotnet", "dotnet.target_framework", "net10.0", "The recipe is valid."},
			[]string{"deploy", "net8.0"}},
		{[]string{"--choice", "language=csharp", "--choice", "kind=service"},
			[]string{"language     dotnet", "deploy       docker", "kind         service        (answered"},
			nil},
		{[]string{"--choice", "language=node"},
			[]string{"language     node", "lint         eslint", "test         vitest", "eslint.config", "@monosmith/eslint-config", "(recipe company)", "The recipe is valid."},
			[]string{"structure"}},
	} {
		out := w.smith(true, append([]string{"recipe", "check", "team.recipe.toml"}, tc.args...)...)
		for _, want := range tc.want {
			if !strings.Contains(out, want) {
				t.Errorf("recipe check %v must print %q:\n%s", tc.args, want, out)
			}
		}
		for _, not := range tc.not {
			if strings.Contains(out, not) {
				t.Errorf("recipe check %v must not print %q:\n%s", tc.args, not, out)
			}
		}
	}

	// The language has no default, so the choice must be answered, and the
	// error says how.
	if out := w.smith(false, "recipe", "check", "team.recipe.toml"); !strings.Contains(out, "--choice language=python|csharp|node") || !strings.Contains(out, "needs an answer") {
		t.Fatalf("an unanswered choice:\n%s", out)
	}
	if out := w.smith(false, "recipe", "check", "team.recipe.toml", "--choice", "language=cobol"); !strings.Contains(out, `"cobol" is not one of the options of choice "language" (python, csharp, node)`) {
		t.Fatalf("an answer that is not an option:\n%s", out)
	}

	// A set of plugins that cannot work together is refused with the lines
	// that chose them: ESLint applies to Node projects only.
	w.write("broken.recipe.toml", "name = \"broken\"\n\n[use]\nlanguage = \"python\"\nlint = \"eslint\"\n")
	out := w.smith(false, "recipe", "check", "broken.recipe.toml")
	for _, want := range []string{"eslint applies to [node], not to python projects", "lint = \"eslint\"", "broken.recipe.toml:5 (recipe broken)"} {
		if !strings.Contains(out, want) {
			t.Errorf("a recipe that cannot work must say %q:\n%s", want, out)
		}
	}
	w.write("typo.recipe.toml", "name = \"typo\"\nlangauge = \"python\"\n")
	if out := w.smith(false, "recipe", "check", "typo.recipe.toml"); !strings.Contains(out, `typo.recipe.toml:2:1: unknown key "langauge"`) {
		t.Fatalf("an unknown key:\n%s", out)
	}

	// Tools read the same through --json.
	var report struct {
		Name    string
		Chain   []struct{ Name string }
		Choices []struct{ Name, Answer string }
		Use     []struct{ Key string }
		Errors  []string
	}
	if err := json.Unmarshal([]byte(w.smith(true, "recipe", "check", "team.recipe.toml", "--choice", "language=python", "--json")), &report); err != nil {
		t.Fatal(err)
	}
	if report.Name != "team" || len(report.Chain) != 2 || len(report.Choices) != 2 || len(report.Use) != 2 || len(report.Errors) != 0 {
		t.Errorf("report = %+v", report)
	}
	if err := json.Unmarshal([]byte(w.smith(false, "recipe", "check", "broken.recipe.toml", "--json")), &report); err != nil || len(report.Errors) == 0 {
		t.Errorf("a recipe that is not valid must say why in the JSON: %v %+v", err, report)
	}
}
