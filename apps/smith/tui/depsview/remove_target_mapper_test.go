package depsview

import (
	"reflect"
	"testing"
)

func TestARootOnlyPackageIsRemovedWithTheRootTarget(t *testing.T) {
	m := FromResult(Result{Rows: []Row{{Name: "@swc/core", Projects: []string{"(root)"}}}})
	name, got, ok := m.RemoveTargets()
	if !ok || name != "@swc/core" || !reflect.DeepEqual(got, []Target{{Label: "(root)", Value: "."}}) {
		t.Fatalf("got %q %+v %v", name, got, ok)
	}
}

func TestEachDeclaringProjectIsListedOnce(t *testing.T) {
	m := FromResult(Result{Rows: []Row{{Name: "zod", Projects: []string{"web", "(root)", "web"}}}})
	_, got, _ := m.RemoveTargets()
	want := []Target{{Label: "web", Value: "web"}, {Label: "(root)", Value: "."}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestNoRowMeansNoTarget(t *testing.T) {
	if _, _, ok := FromResult(Result{}).RemoveTargets(); ok {
		t.Fatal("an empty view has nothing to remove")
	}
}
