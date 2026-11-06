package operationplan

import (
	"strings"
	"testing"
)

func planOf(lines ...Line) Plan {
	p := Plan{Title: "t"}
	for _, l := range lines {
		p.Add("g", l)
	}
	return p
}

func TestRisky(t *testing.T) {
	for _, tc := range []struct {
		name string
		plan Plan
		want bool
	}{
		{"files created only", planOf(Line{Action: Create, Path: "a"}, Line{Action: Change, Path: "smith.toml"}), false},
		{"a deletion", planOf(Line{Action: Remove, Path: "a"}), true},
		{"an edit shown with a diff", planOf(Line{Action: Change, Path: "a", Diff: "-x\n+y"}), true},
		{"an edit of a file the person owns", planOf(Line{Action: Change, Path: "a", Owner: "you"}), true},
		{"an edit of a file smith owns", planOf(Line{Action: Change, Path: "a", Owner: "smith"}), false},
		{"nothing", Plan{}, false},
	} {
		if got := tc.plan.Risky(); got != tc.want {
			t.Errorf("%s: got %v", tc.name, got)
		}
	}
}

// confirm = "risky" applies a plan that cannot lose anything, says so, and still
// asks for a risky one. The default asks for both.
func TestPresentHonoursTheConfirmSetting(t *testing.T) {
	safe := planOf(Line{Action: Create, Path: "a"})
	risky := planOf(Line{Action: Remove, Path: "a"})
	nobody := Environment{}
	for _, tc := range []struct {
		name    string
		plan    Plan
		confirm string
		env     Environment
		input   string
		want    bool
		wantErr bool
		says    string
	}{
		{"always asks for a safe plan: no terminal refuses", safe, "", nobody, "", false, true, ""},
		{"always asks for a safe plan: the answer decides", safe, "always", Environment{Terminal: true}, "y\n", true, false, ""},
		{"risky applies a safe plan without a terminal", safe, ConfirmRisky, nobody, "", true, false, "Applying it without asking"},
		{"risky asks for a risky plan: no terminal refuses", risky, ConfirmRisky, nobody, "", false, true, ""},
		{"risky asks for a risky plan: no answers no", risky, ConfirmRisky, Environment{Terminal: true}, "n\n", false, true, ""},
		{"risky asks for a risky plan: yes goes on", risky, ConfirmRisky, Environment{Terminal: true}, "y\n", true, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			got, err := Present(&out, strings.NewReader(tc.input), tc.plan, Options{Ask: true, Env: tc.env, Confirm: tc.confirm})
			if got != tc.want || (err != nil) != tc.wantErr {
				t.Fatalf("got %v, %v\n%s", got, err, out.String())
			}
			if tc.says != "" && !strings.Contains(out.String(), tc.says) {
				t.Errorf("the output should say %q:\n%s", tc.says, out.String())
			}
		})
	}
}
