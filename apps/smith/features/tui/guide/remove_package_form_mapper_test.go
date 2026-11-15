package guide

import (
	"reflect"
	"testing"
)

func TestRemovingAPackageNamesEveryProjectPicked(t *testing.T) {
	a := RemovePackageAction("zod", []RemoveTarget{{"web", "web"}, {"(root)", "."}})
	if len(a.Steps) != 1 || !a.Steps[0].Multi || len(a.Steps[0].Options) != 2 {
		t.Fatalf("one multiple choice of the two projects: %+v", a.Steps)
	}
	got := a.Argv(map[string][]string{"projects": {"web", "."}})
	if want := []string{"uninstall", "-p", "web", "-p", ".", "zod"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestNothingRunsWhileNoProjectIsPicked(t *testing.T) {
	if got := RemovePackageArgv("zod", map[string][]string{"projects": {""}}); got != nil {
		t.Fatalf("got %v", got)
	}
}
