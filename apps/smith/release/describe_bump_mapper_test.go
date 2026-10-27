package release

import "testing"

func TestWhyNamesTheCommitThatDecidedTheBump(t *testing.T) {
	for name, tc := range map[string]struct {
		item Item
		want string
	}{
		"feature over fixes": {Item{Changes: []Change{{Type: "fix", Description: "handle empty input"}, {Type: "feat", Description: "add search"}, {Type: "fix", Description: "x"}}}, "feat: add search (+2 more)"},
		"breaking wins":      {Item{Changes: []Change{{Type: "feat", Description: "add search"}, {Type: "refactor", Scopes: []string{"api"}, Description: "drop v1", Breaking: true}}}, "refactor(api)!: drop v1 (+1 more)"},
		"a single fix":       {Item{Changes: []Change{{Type: "fix", Description: "handle empty input"}}}, "fix: handle empty input"},
		"a dependency":       {Item{Reason: "dependency core releases"}, "dependency core releases"},
		"forced":             {Item{Forced: true, Reason: "forced by version(core)[2.0.0]", Changes: []Change{{Type: "fix", Description: "x"}}}, "forced by version(core)[2.0.0]"},
		"a long description": {Item{Changes: []Change{{Type: "feat", Description: "add a search that looks through every single package in the whole workspace at once"}}}, "feat: add a search that looks through every single package in…"},
	} {
		if got := tc.item.Why(); got != tc.want {
			t.Errorf("%s: %q, want %q", name, got, tc.want)
		}
	}
}
