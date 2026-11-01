package tui

import (
	"strings"
	"testing"
)

func TestTheAddPlanNamesTheChoicesAndTheSiblings(t *testing.T) {
	for _, tc := range []struct {
		argv     []string
		projects int
		want     []string
		not      []string
	}{
		{[]string{"add", "app", "web", "--with", "react,vitest"}, 4, []string{"your choices: react, vitest", "follows the other projects"}, []string{"defaults"}},
		{[]string{"add", "app", "web", "--fresh"}, 4, []string{"stack's own defaults", "does not follow the other projects"}, []string{"follows the other"}},
		{[]string{"add", "app", "web"}, 1, []string{"stack's own defaults"}, []string{"follow"}},
	} {
		got := addPlanWhy(tc.argv, tc.projects)
		for _, w := range tc.want {
			if !strings.Contains(got, w) {
				t.Errorf("%v: missing %q in %q", tc.argv, w, got)
			}
		}
		for _, n := range tc.not {
			if strings.Contains(got, n) {
				t.Errorf("%v: unexpected %q in %q", tc.argv, n, got)
			}
		}
	}
}
