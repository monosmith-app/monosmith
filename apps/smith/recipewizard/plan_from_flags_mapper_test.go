package recipewizard

import (
	"strings"
	"testing"
)

func TestPlanFromFlagsFixesEachLevel(t *testing.T) {
	p, err := PlanFromFlags("team", "", "base.recipe.toml", []string{"language=python", "structure=slices, layers"})
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
			_, err := PlanFromFlags("team", "1.0.0", "", tc.fixed)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to say %q", err, tc.want)
			}
		})
	}
	if _, err := PlanFromFlags("Bad Name", "1.0.0", "", nil); err == nil {
		t.Error("a bad name was accepted")
	}
}
