package projectgraph

import (
	"reflect"
	"testing"
)

func TestCycles(t *testing.T) {
	ps := []*Project{
		proj("web", "apps/web", []string{"web"}, []string{"ui"}),
		proj("ui", "libs/ui", []string{"ui"}, []string{"core", "web"}),
		proj("core", "libs/core", []string{"core"}, []string{"ui"}),
		proj("api", "apps/api", []string{"api"}, []string{"db"}),
		proj("db", "libs/db", []string{"db"}, []string{"api"}),
		proj("leaf", "libs/leaf", []string{"leaf"}, nil),
	}
	ResolveEdges(ps)
	g, err := New(ps)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"api", "db", "api"}, {"core", "ui", "core"}}
	if got := g.Cycles(); !reflect.DeepEqual(got, want) {
		t.Fatalf("cycles %v, want %v", got, want)
	}
	if got := sample(t).Cycles(); len(got) != 0 {
		t.Fatalf("an acyclic graph has none: %v", got)
	}
}

func TestCycleRemedyNamesTheEdgeThatClosesTheLoop(t *testing.T) {
	cases := map[string][]string{
		"remove web's import of ui (web → ui)": {"ui", "web", "ui"},
		"remove a's import of a (a → a)":       {"a", "a"},
		"":                                     {"a"},
	}
	for want, cycle := range cases {
		if got := CycleRemedy(cycle); got != want {
			t.Errorf("CycleRemedy(%v) = %q, want %q", cycle, got, want)
		}
	}
}
