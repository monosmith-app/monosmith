package recipewizard

import (
	"strings"
	"testing"
)

func TestPlanFromFlagsFixesEachLevel(t *testing.T) {
	p, err := PlanFromFlags("team", "", "base.recipe.toml", []string{"language=python", "structure=slices, layers"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Version != "0.1.0" || p.Extends != "base.recipe.toml" || p.Levels["language"].Fixed[0] != "python" || len(p.Levels["structure"].Fixed) != 2 {
		t.Errorf("plan = %+v", p)
	}
}

func TestPlanFromFlagsRefusesWhatARecipeCouldNotUse(t *testing.T) {
	for name, tc := range map[string]struct {
		fixed []string
		want  string
	}{
		"no equals":    {[]string{"python"}, "role=plugin"},
		"empty plugin": {[]string{"language="}, "role=plugin"},
		"twice":        {[]string{"language=python", "language=node"}, "twice"},
		"two for one":  {[]string{"language=python,node"}, "takes one plugin"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := PlanFromFlags("team", "1.0.0", "", tc.fixed, nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to say %q", err, tc.want)
			}
		})
	}
	if _, err := PlanFromFlags("Bad Name", "1.0.0", "", nil, nil); err == nil {
		t.Error("a bad name was accepted")
	}
}

func TestPlanFromFlagsChosenLevel(t *testing.T) {
	p, err := PlanFromFlags("team", "", "", []string{"scaffold=none"}, []string{
		"language=Which language?:python=python,csharp=dotnet",
		"structure=Which layout?:small=slices,big=slices+layers",
	})
	if err != nil {
		t.Fatal(err)
	}
	lang := p.Levels["language"]
	if lang.Choice != "language" || lang.Plugins["python"] != "python" || lang.Plugins["csharp"] != "dotnet" {
		t.Errorf("language level = %+v", lang)
	}
	if c := p.Choices["language"]; c.Ask != "Which language?" || strings.Join(c.Options, ",") != "python,csharp" || c.Default != "python" {
		t.Errorf("choice = %+v", c)
	}
	if got := p.Levels["structure"].PluginsOf("big"); strings.Join(got, "+") != "slices+layers" {
		t.Errorf("structure for big = %v", got)
	}
}

func TestPlanFromFlagsRefusesBadChosen(t *testing.T) {
	for name, tc := range map[string]struct {
		fixed, chosen []string
		want          string
	}{
		"no question":      {nil, []string{"language=python=python,csharp=dotnet"}, "must be level=question"},
		"no equals":        {nil, []string{"language"}, "must be level=question"},
		"empty plugin":     {nil, []string{"language=Which?:python=,csharp=dotnet"}, "must be level=question"},
		"same answer":      {nil, []string{"language=Which?:python=python,python=node"}, "twice"},
		"fixed and chosen": {[]string{"language=python"}, []string{"language=Which?:python=python,csharp=dotnet"}, "fixed or chosen"},
		"one answer":       {nil, []string{"language=Which?:python=python"}, "at least two options"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := PlanFromFlags("team", "1.0.0", "", tc.fixed, tc.chosen)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to say %q", err, tc.want)
			}
		})
	}
}
