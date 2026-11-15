package removalplan

import (
	"strings"
	"testing"
)

func TestRaiseESLint(t *testing.T) {
	for _, tc := range []struct {
		name, manifest string
		raised         string
		problem        string
		keeps          string
	}{
		{
			name:     "eslint 9 in devDependencies and a literal override",
			manifest: `{"devDependencies": {"eslint": "^9.39.3"}, "overrides": {"eslint": "~9.0.0"}}`,
			raised:   "^9.39.3,~9.0.0",
		},
		{
			name:     "a $eslint override follows the dependency and stays",
			manifest: `{"devDependencies": {"eslint": "^9.39.3"}, "overrides": {"eslint": "$eslint"}}`,
			raised:   "^9.39.3",
			keeps:    `"eslint": "$eslint"`,
		},
		{
			name:     "an ESLint the config runs with is left alone",
			manifest: `{"devDependencies": {"eslint": "^10.12.0"}}`,
			keeps:    `"eslint": "^10.12.0"`,
		},
		{
			name:     "a compound range is a range: raised",
			manifest: `{"dependencies": {"eslint": ">=8 <10"}}`,
			raised:   ">=8 <10",
		},
		{
			name:     "a tag or a path cannot be raised: a problem",
			manifest: `{"devDependencies": {"eslint": "latest"}}`,
			problem:  `devDependencies name eslint "latest"`,
			keeps:    `"eslint": "latest"`,
		},
		{
			name:     "a peer range is never edited, only reported",
			manifest: `{"peerDependencies": {"eslint": "^9.0.0"}}`,
			problem:  `peerDependencies ask for eslint "^9.0.0"`,
			keeps:    `"eslint": "^9.0.0"`,
		},
		{
			name:     "no eslint, nothing to do",
			manifest: `{"devDependencies": {"jest": "^30"}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			next, raised, problem, err := RaiseESLint(tc.manifest)
			if err != nil {
				t.Fatal(err)
			}
			if got := strings.Join(raised, ","); got != tc.raised {
				t.Errorf("raised %q, want %q", got, tc.raised)
			}
			if tc.raised != "" && !strings.Contains(next, `"eslint": "`+LintESLintRange+`"`) {
				t.Errorf("not raised to %s:\n%s", LintESLintRange, next)
			}
			if tc.raised == "" && next != tc.manifest {
				t.Errorf("changed though nothing was raised:\n%s", next)
			}
			if (tc.problem == "") != (problem == "") || !strings.Contains(problem, tc.problem) {
				t.Errorf("problem %q, want %q", problem, tc.problem)
			}
			if tc.keeps != "" && !strings.Contains(next, tc.keeps) {
				t.Errorf("%s is gone:\n%s", tc.keeps, next)
			}
		})
	}
}
