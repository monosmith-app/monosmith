package convert

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
)

func TestConfirmationTextNamesWhatIsDeletedAndTheNpmSwitch(t *testing.T) {
	cases := []struct {
		name     string
		plan     convertplan.Plan
		want     []string
		wantNot  []string
		switches bool
	}{
		{"nx on npm", convertplan.Plan{Delete: []string{"nx.json", ".nx"}}, []string{"delete 2 file(s)", "nx.json", ".nx", "regenerate the CI"}, []string{"SWITCH"}, false},
		{"pnpm", convertplan.Plan{Delete: []string{"pnpm-lock.yaml"}, SetWorkspaces: []string{"apps/*"}}, []string{"SWITCH", "pnpm-lock.yaml", "rewrite the root package.json"}, nil, true},
		{"yarn", convertplan.Plan{Delete: []string{"yarn.lock"}}, []string{"SWITCH this pnpm or yarn workspace to npm"}, nil, true},
	}
	for _, c := range cases {
		plan := c.plan
		got := ConfirmationText(&plan)
		for _, w := range c.want {
			if !strings.Contains(got, w) {
				t.Errorf("%s: missing %q in:\n%s", c.name, w, got)
			}
		}
		for _, w := range c.wantNot {
			if strings.Contains(got, w) {
				t.Errorf("%s: unexpected %q", c.name, w)
			}
		}
		if SwitchesToNpm(&plan) != c.switches {
			t.Errorf("%s: SwitchesToNpm = %v", c.name, !c.switches)
		}
	}
}

func TestConfirmationTextCapsTheDeletedList(t *testing.T) {
	plan := convertplan.Plan{}
	for i := 0; i < 20; i++ {
		plan.Delete = append(plan.Delete, "f"+string(rune('a'+i)))
	}
	if got := ConfirmationText(&plan); !strings.Contains(got, "and 8 more") {
		t.Fatalf("%s", got)
	}
}
