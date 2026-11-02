package catalog

import "testing"

func TestPlanFlag(t *testing.T) {
	for _, tc := range []struct {
		argv []string
		want string
	}{
		{[]string{"upgrade"}, "--dry-run"},
		{[]string{"ci", "generate"}, "--dry-run"},
		{[]string{"sync"}, "--dry-run"},
		{[]string{"release"}, ""}, // the shell reads the release plan itself
		{[]string{"deps", "up", "--yes"}, "--check"},
		{[]string{"bootstrap"}, "--dry-run"},
		{[]string{"remove", "web", "--keep-copy"}, "--dry-run"},
		{[]string{"upgrade", "--dry-run"}, ""},  // already a plan
		{[]string{"sync", "--check"}, ""},       // reads only
		{[]string{"doctor"}, ""},                // reads only
		{[]string{"plugin", "enable", "x"}, ""}, // no plan flag
		{[]string{"nothing-like-this"}, ""},     // unknown commands have none
	} {
		got, ok := PlanFlag(tc.argv)
		if got != tc.want || ok != (tc.want != "") {
			t.Errorf("%v: got %q (%v), want %q", tc.argv, got, ok, tc.want)
		}
	}
}
