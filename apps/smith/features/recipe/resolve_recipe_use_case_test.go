package recipe

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func chain(docs ...Document) Loaded { return Loaded{Chain: docs} }

func usesOf(r Resolution) map[string]string {
	out := map[string]string{}
	for _, u := range r.Uses {
		out[u.Key] = strings.Join(u.Plugins, ",")
	}
	return out
}

func TestResolveAppliesTheBlocksOfTheAnswers(t *testing.T) {
	team := mustParse(t, "team.toml", teamRecipe)

	py, err := Resolve(chain(team), map[string]string{"language": "python"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"structure": "vertical-slices", "language": "python"}
	if !reflect.DeepEqual(usesOf(py), want) {
		t.Errorf("python uses = %v, want %v", usesOf(py), want)
	}
	if u, _ := py.UseFor("language"); u.Origin.Line != 17 || u.Origin.File != "team.toml" || u.Origin.Recipe != "team" {
		t.Errorf("language origin = %+v", u.Origin)
	}
	// The kind choice sits at the top of the recipe, so it is asked for Python
	// too, and answered by its default.
	if len(py.Asked) != 2 || py.Asked[1].Name != "kind" || py.Asked[1].Answer != "library" || py.Asked[1].Given {
		t.Errorf("asked = %+v", py.Asked)
	}

	svc, err := Resolve(chain(team), map[string]string{"language": "csharp", "kind": "service"})
	if err != nil {
		t.Fatal(err)
	}
	want = map[string]string{"structure": "vertical-slices", "language": "dotnet", "deploy": "docker"}
	if !reflect.DeepEqual(usesOf(svc), want) {
		t.Errorf("csharp service uses = %v, want %v", usesOf(svc), want)
	}
	if svc.Settings["dotnet"].(map[string]any)["target_framework"] != "net10.0" {
		t.Errorf("settings = %v", svc.Settings)
	}
	if o := svc.Origins["settings.dotnet.target_framework"]; o.Line != 23 {
		t.Errorf("setting origin = %+v", o)
	}

	lib, err := Resolve(chain(team), map[string]string{"language": "csharp"})
	if err != nil {
		t.Fatal(err)
	}
	if _, has := lib.UseFor("deploy"); has {
		t.Error("the service block must not apply to a library")
	}
}

func TestResolveMergesTheChainOldestFirst(t *testing.T) {
	parent := mustParse(t, "company.toml", `
name = "company"
version = "3.0.0"

[choices.language]
options = ["python", "csharp"]
default = "python"

[use]
language = "python"
structure = "vertical-slices"

[settings.ruff]
line_length = 100

[settings.ruff.format]
quote = "double"
`)
	child := mustParse(t, "team.toml", `
name = "team"
version = "1.0.0"
extends = "company.toml"

[choices.language]
default = "csharp"

[use]
structure = "other-structure"

[settings.ruff]
line_length = 88

[settings.ruff.format]
indent = 4
`)
	r, err := Resolve(chain(parent, child), nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != "team" || r.Version != "1.0.0" || len(r.Chain) != 2 || r.Chain[0].Name != "company" {
		t.Errorf("resolution = %+v", r)
	}
	// The child changes the default only; the options stay the parent's.
	if q := r.Asked[0]; q.Answer != "csharp" || len(q.Options) != 2 || q.Origin.Recipe != "team" {
		t.Errorf("choice = %+v", q)
	}
	if got := usesOf(r); got["structure"] != "other-structure" || got["language"] != "python" {
		t.Errorf("uses = %v", got)
	}
	if u, _ := r.UseFor("language"); u.Origin.Recipe != "company" {
		t.Errorf("the parent's language keeps its origin: %+v", u.Origin)
	}
	ruff := r.Settings["ruff"].(map[string]any)
	format := ruff["format"].(map[string]any)
	if ruff["line_length"] != int64(88) || format["quote"] != "double" || format["indent"] != int64(4) {
		t.Errorf("settings = %v", r.Settings)
	}
	for path, want := range map[string]Origin{
		"settings.ruff.line_length":   {Recipe: "team", File: "team.toml", Line: 12},
		"settings.ruff.format.quote":  {Recipe: "company", File: "company.toml", Line: 16},
		"settings.ruff.format.indent": {Recipe: "team", File: "team.toml", Line: 15},
	} {
		if got := r.Origins[path]; got != want {
			t.Errorf("origin of %s = %+v, want %+v", path, got, want)
		}
	}
}

func TestResolveAppliesMatchingBlocksInFileOrder(t *testing.T) {
	// when.b comes before when.a in the file, so a wins; alphabetical order
	// would give b.
	d := mustParse(t, "r.toml", `
name = "r"

[choices.a]
options = ["x"]
default = "x"

[choices.b]
options = ["y"]
default = "y"

[when.b.y.use]
lint = "from-b"

[when.a.x.use]
lint = "from-a"
`)
	r, err := Resolve(chain(d), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := usesOf(r)["lint"]; got != "from-a" {
		t.Errorf("lint = %q, the block later in the file must win", got)
	}
}

func TestResolveAsksABlocksOwnChoicesOnlyWhenItMatches(t *testing.T) {
	d := mustParse(t, "r.toml", `
name = "r"

[choices.language]
options = ["python", "csharp"]
default = "python"

[when.language.csharp.choices.framework]
options = ["net8", "net10"]
default = "net10"

[when.language.csharp.settings.dotnet]
note = "csharp"

[when.language.csharp.when.framework.net10.use]
deploy = "docker"
`)
	py, err := Resolve(chain(d), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(py.Asked) != 1 {
		t.Errorf("python must not be asked the framework: %+v", py.Asked)
	}
	cs, err := Resolve(chain(d), map[string]string{"language": "csharp", "framework": "net10"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cs.Asked) != 2 || cs.Asked[1].Name != "framework" || !cs.Asked[1].Given || usesOf(cs)["deploy"] != "docker" {
		t.Errorf("asked = %+v, uses = %v", cs.Asked, usesOf(cs))
	}
	// The answer to a choice that is not asked is a mistake worth saying.
	_, err = Resolve(chain(d), map[string]string{"framework": "net8"})
	if err == nil || !strings.Contains(err.Error(), `asks no choice named "framework"`) {
		t.Errorf("err = %v", err)
	}
}

func TestResolveNamesEveryChoiceWithNoAnswer(t *testing.T) {
	d := mustParse(t, "team.toml", teamRecipe)
	r, err := Resolve(chain(d), nil)
	var missing *MissingAnswerError
	if !errors.As(err, &missing) || len(missing.Choices) != 1 || missing.Choices[0].Name != "language" {
		t.Fatalf("err = %v", err)
	}
	for _, want := range []string{`choice "language" (team.toml:4 (recipe team))`, "--choice language=python|csharp"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%q must contain %q", err, want)
		}
	}
	// What was asked is still reported, for a caller that wants to prompt.
	if len(r.Asked) != 2 || r.Asked[0].Answer != "" {
		t.Errorf("asked = %+v", r.Asked)
	}
}

func TestResolveRefusesWhatDoesNotAddUp(t *testing.T) {
	d := mustParse(t, "team.toml", teamRecipe)
	for name, tc := range map[string]struct {
		d       Document
		answers map[string]string
		want    string
	}{
		"answer not an option": {d, map[string]string{"language": "cobol"}, `"cobol" is not one of the options of choice "language" (python, csharp)`},
		"when names an undeclared choice": {mustParse(t, "r.toml", "name = \"r\"\n[when.nope.x.use]\nlint = \"a\"\n"),
			nil, `r.toml:2: when.nope refers to a choice no recipe declares`},
		"when names a typo of an answer": {mustParse(t, "r.toml", "name = \"r\"\n[choices.a]\noptions = [\"x\"]\ndefault = \"x\"\n[when.a.y.use]\nlint = \"a\"\n"),
			nil, `r.toml:5: when.a.y: "y" is not one of the options of choice "a" (x)`},
		"a choice declared again after it was asked": {mustParse(t, "r.toml", "name = \"r\"\n[choices.a]\noptions = [\"x\"]\ndefault = \"x\"\n[when.a.x.choices.a]\ndefault = \"x\"\n"),
			nil, `choice "a" was already asked and cannot be declared again here`},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Resolve(chain(tc.d), tc.answers)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
	if _, err := Resolve(Loaded{}, nil); err == nil {
		t.Error("an empty chain must be refused")
	}
}
