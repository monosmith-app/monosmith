package prompt

import (
	"slices"
	"testing"
)

func TestAffectedByAddsDependentsAndPicksTheDeepestRoot(t *testing.T) {
	projects := []Project{
		{Name: "core", Root: "libs/core", Dependents: []string{"api"}},
		{Name: "api", Root: "apps/api"},
		{Name: "outer", Root: "tools"},
		{Name: "inner", Root: "tools/inner"},
	}
	got := AffectedBy(projects, []string{"libs/core/a.go", "tools/inner/b.go", "README.md", "libs/corex/c.go"})
	want := []string{"api", "core", "inner"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAffectedByIsEmptyWhenNothingChanged(t *testing.T) {
	if got := AffectedBy([]Project{{Name: "a", Root: "a"}}, nil); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}
