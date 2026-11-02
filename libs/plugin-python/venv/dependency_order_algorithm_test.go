package venv

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func projectNames(ps []protocol.Project) string {
	s := ""
	for _, p := range ps {
		s += p.Name + " "
	}
	return s
}

func TestDependencyOrderPutsALibraryBeforeTheProjectThatNeedsIt(t *testing.T) {
	in := []protocol.Project{
		{Name: "api", Exports: []string{"api"}, Imports: []string{"core_py", "requests"}},
		{Name: "core-py", Exports: []string{"core_py"}},
		{Name: "tool", Exports: []string{"tool"}},
	}
	if got := projectNames(DependencyOrder(in)); got != "core-py api tool " {
		t.Fatalf("got %q", got)
	}
}

func TestDependencyOrderDoesNotLoopOnACycle(t *testing.T) {
	in := []protocol.Project{
		{Name: "a", Exports: []string{"a"}, Imports: []string{"b"}},
		{Name: "b", Exports: []string{"b"}, Imports: []string{"a"}},
	}
	if got := projectNames(DependencyOrder(in)); got != "b a " {
		t.Fatalf("got %q", got)
	}
}
