package convert

import (
	"bufio"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func outputChoicePlan() *convertplan.Plan {
	plan := &convertplan.Plan{Source: convertplan.SourceNx}
	plan.SetTask("app", "build", workspace.TaskOverride{Outputs: []string{"dist"}})
	plan.OutputChoices = []convertplan.OutputChoice{{Project: "app", Task: "build", Declared: []string{"dist", "dist-dev"}, Detected: []string{"dist"}}}
	return plan
}

func TestAskOutputChoices(t *testing.T) {
	cases := []struct {
		answer string
		want   []string
		kept   bool
	}{
		{"\n", []string{"dist"}, false},
		{"1\n", []string{"dist"}, false},
		{"2\n", []string{"dist", "dist-dev"}, true},
	}
	for _, c := range cases {
		plan := outputChoicePlan()
		var out strings.Builder
		if err := AskOutputChoices(bufio.NewReader(strings.NewReader(c.answer)), &out, plan, nil); err != nil {
			t.Fatal(err)
		}
		if got := plan.Config.Projects["app"].Tasks["build"].Outputs; !slices.Equal(got, c.want) || plan.OutputChoices[0].Kept != c.kept {
			t.Errorf("answer %q: outputs %v kept %v, want %v %v", c.answer, got, plan.OutputChoices[0].Kept, c.want, c.kept)
		}
		if !strings.Contains(out.String(), "app:build writes dist, but Nx declared dist, dist-dev as its outputs.") || !strings.Contains(out.String(), "1) dist, what the build writes (suggested)") {
			t.Errorf("question:\n%s", out.String())
		}
	}
	if err := AskOutputChoices(bufio.NewReader(strings.NewReader("3\n")), &strings.Builder{}, outputChoicePlan(), nil); err == nil {
		t.Error("an answer outside the options is an error")
	}
}
