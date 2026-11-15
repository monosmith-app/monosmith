package dependencies

import (
	"reflect"
	"testing"
)

func TestTheWorkspaceRootIsNamedByADotAndTakenOutOfTheTargets(t *testing.T) {
	root, others := SplitRootTarget([]string{"web", ".", "api"})
	if !root || !reflect.DeepEqual(others, []string{"web", "api"}) {
		t.Fatalf("got %v %v", root, others)
	}
	if root, others = SplitRootTarget([]string{"web"}); root || len(others) != 1 {
		t.Fatalf("no root named: %v %v", root, others)
	}
}

func TestRootInstallCommand(t *testing.T) {
	if got := RootInstallCommand([]string{"vitest", "zod@3"}, true); got != "npm install vitest zod@3 --save-dev" {
		t.Fatalf("got %q", got)
	}
	if got := RootInstallCommand([]string{"zod"}, false); got != "npm install zod" {
		t.Fatalf("got %q", got)
	}
}
