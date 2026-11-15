package smithfix

import (
	"slices"
	"testing"
)

func TestStepAddsYesWhereTheCommandAsksOnATerminal(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want []string
	}{
		{[]string{"upgrade", "--defaults"}, []string{"smith", "upgrade", "--defaults", "--yes"}},
		{[]string{"derive"}, []string{"smith", "derive", "--yes"}},
		{[]string{"cache", "prune"}, []string{"smith", "cache", "prune"}},
	} {
		if got := Step("x", tc.args...).Argv; !slices.Equal(got, tc.want) {
			t.Errorf("%v: got %v, want %v", tc.args, got, tc.want)
		}
	}
}

func TestPlanArgvIsTheCommandWithItsPlanFlagAndWithoutYes(t *testing.T) {
	for _, tc := range []struct {
		argv []string
		want []string
		ok   bool
	}{
		{[]string{"smith", "upgrade", "--defaults", "--yes"}, []string{"smith", "upgrade", "--defaults", "--dry-run"}, true},
		{[]string{"smith", "ci", "generate", "--yes"}, []string{"smith", "ci", "generate", "--dry-run"}, true},
		{[]string{"smith", "agent", "install", "--yes"}, nil, false}, // no plan flag
		{[]string{"npm", "install"}, nil, false},                     // not smith
		{[]string{"smith"}, nil, false},
	} {
		got, ok := PlanArgv(tc.argv)
		if ok != tc.ok || !slices.Equal(got, tc.want) {
			t.Errorf("%v: got %v (%v), want %v (%v)", tc.argv, got, ok, tc.want, tc.ok)
		}
	}
}

func TestAFixOfSmithCommandsIsNeverSafe(t *testing.T) {
	f := Fix("move the defaults", Step("move", "upgrade", "--defaults"))
	c := Choice("settle", "Which?", Option("Replace", "derive", "--replace", "a"))
	if f.Safe || c.Safe || len(c.Choices[0].Options) != 1 {
		t.Fatalf("%+v %+v", f, c)
	}
}
