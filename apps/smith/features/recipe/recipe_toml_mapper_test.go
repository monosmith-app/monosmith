package recipe

import (
	"strings"
	"testing"
)

// mustParse parses a recipe whose text starts on its first line.
func mustParse(t *testing.T, ref, text string) Document {
	t.Helper()
	d, err := ParseRecipe([]byte(strings.TrimPrefix(text, "\n")), Source{Ref: ref})
	if err != nil {
		t.Fatalf("%s: %v", ref, err)
	}
	return d
}

const teamRecipe = `
name = "team"
version = "1.2.0"

[choices.language]
ask = "Language"
options = ["python", "csharp"]

[choices.kind]
ask = "What are you creating?"
options = ["library", "service"]
default = "library"

[use]
structure = "vertical-slices"

[when.language.python.use]
language = "python"

[when.language.csharp.use]
language = "dotnet"

[when.language.csharp.settings.dotnet]
target_framework = "net10.0"

[when.language.csharp.when.kind.service.use]
deploy = "docker"
`

func TestParseRecipeReadsEveryPartAndItsLines(t *testing.T) {
	d := mustParse(t, "team.toml", teamRecipe)
	r := d.Recipe
	if r.Name != "team" || r.Version != "1.2.0" || r.Extends != "" {
		t.Errorf("recipe = %+v", r)
	}
	if c := r.Choices["kind"]; c.Ask != "What are you creating?" || c.Default != "library" || len(c.Options) != 2 {
		t.Errorf("choice = %+v", c)
	}
	csharp := r.When["language"]["csharp"]
	if csharp.Use["language"] != "dotnet" || csharp.Settings["dotnet"]["target_framework"] != "net10.0" {
		t.Errorf("csharp block = %+v", csharp)
	}
	if csharp.When["kind"]["service"].Use["deploy"] != "docker" {
		t.Errorf("nested block = %+v", csharp.When)
	}
	for path, want := range map[string]int{
		"name":                                 1,
		"choices.language":                     4,
		"use.structure":                        14,
		"when.language.python":                 16,
		"when.language.python.use.language":    17,
		"when.language.csharp.settings.dotnet": 22,
		"when.language.csharp.settings.dotnet.target_framework": 23,
		"when.language.csharp.when.kind.service":                25,
	} {
		if got := d.Line(path); got != want {
			t.Errorf("Line(%q) = %d, want %d", path, got, want)
		}
	}
}

func TestParseRecipeFindsKeysInInlineTables(t *testing.T) {
	d := mustParse(t, "r.toml", `
name = "r"
use = { language = "go" }
settings = { eslint = { config = "@acme/eslint-config" } }
`)
	if d.Recipe.Use["language"] != "go" || d.Line("use.language") != 2 || d.Line("settings.eslint.config") != 3 {
		t.Errorf("use = %v, lines %d %d", d.Recipe.Use, d.Line("use.language"), d.Line("settings.eslint.config"))
	}
}

func TestParseRecipeRefusesWhatItDoesNotKnow(t *testing.T) {
	for name, tc := range map[string]struct{ text, want string }{
		"unknown top-level key":     {"name = \"r\"\nusee = {}\n", `r.toml:2:1: unknown key "usee"; here a recipe accepts name, version, extends, choices, use, settings, when`},
		"unknown key in a block":    {"name = \"r\"\n[when.a.b]\nbogus = 1\n", `unknown key "when.a.b.bogus"; here a recipe accepts choices, use, settings, when`},
		"unknown key in a choice":   {"name = \"r\"\n[choices.a]\nask = \"x\"\nbogus = 1\n", `here a recipe accepts ask, options, default`},
		"wrong type":                {"name = 3\n", "r.toml:1:"},
		"not toml":                  {"name = \n", "r.toml:1:"},
		"no name":                   {"version = \"1\"\n", "the recipe has no name"},
		"default not an option":     {"name = \"r\"\n[choices.a]\noptions = [\"x\"]\ndefault = \"y\"\n", `r.toml:2: choice "a" has the default "y", which is not one of its options (x)`},
		"use is not a plugin":       {"name = \"r\"\n[use]\nlanguage = 3\n", "r.toml:3: use.language must be a plugin name, or a list of plugin names"},
		"use is an empty list":      {"name = \"r\"\n[use]\nstructure = []\n", "use.structure must be a plugin name"},
		"bad use in a nested block": {"name = \"r\"\n[when.a.b.use]\nlint = 1\n", "r.toml:3: use.lint must be"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ParseRecipe([]byte(tc.text), Source{Ref: "r.toml"})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestUseListTakesOnePluginOrAList(t *testing.T) {
	if got, ok := UseList("eslint"); !ok || len(got) != 1 || got[0] != "eslint" {
		t.Errorf("one = %v %v", got, ok)
	}
	if got, ok := UseList([]any{"a", "b@^1"}); !ok || len(got) != 2 || got[1] != "b@^1" {
		t.Errorf("list = %v %v", got, ok)
	}
	for _, bad := range []any{"", 3, []any{}, []any{"a", 3}, map[string]any{}} {
		if _, ok := UseList(bad); ok {
			t.Errorf("%v must be refused", bad)
		}
	}
}
