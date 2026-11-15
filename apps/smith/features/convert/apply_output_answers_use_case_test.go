package convert

import (
	"bufio"
	"slices"
	"strings"
	"testing"
)

func TestApplyOutputAnswers(t *testing.T) {
	cases := []struct {
		answer string
		want   []string
		kept   bool
		fails  bool
	}{
		{"app:build=detected", []string{"dist"}, false, false},
		{"app:build=declared", []string{"dist", "dist-dev"}, true, false},
		{"app:build=both", nil, false, true},
		{"web:build=declared", nil, false, true},
		{"app:build", nil, false, true},
		{"build=declared", nil, false, true},
	}
	for _, c := range cases {
		plan := outputChoicePlan()
		given, err := ApplyOutputAnswers(plan, []string{c.answer})
		if c.fails {
			if err == nil {
				t.Errorf("%q: want an error", c.answer)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%q: %v", c.answer, err)
		}
		if got := plan.Config.Projects["app"].Tasks["build"].Outputs; !slices.Equal(got, c.want) || plan.OutputChoices[0].Kept != c.kept || !given["app:build"] {
			t.Errorf("%q: outputs %v kept %v given %v", c.answer, got, plan.OutputChoices[0].Kept, given)
		}
	}
}

func TestAnAnsweredOutputChoiceIsNotAskedAgain(t *testing.T) {
	plan := outputChoicePlan()
	given, err := ApplyOutputAnswers(plan, []string{"app:build=declared"})
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err := AskOutputChoices(bufio.NewReader(strings.NewReader("")), &out, plan, given); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 || !plan.OutputChoices[0].Kept {
		t.Fatalf("asked again:\n%s", out.String())
	}
}
