package projectlist

import (
	"slices"
	"testing"
)

func TestAFilterKeepsTheProjectsThatPassEveryWord(t *testing.T) {
	l := sample()
	for _, tc := range []struct {
		filter string
		want   []string
	}{
		{"", []string{"orders", "web", "kit", "sdk", "zeta"}},
		{"o", []string{"orders", "web", "kit", "sdk", "zeta"}}, // the others match the language or the kind (node, go, python, tool)
		{"kind:app", []string{"orders", "web"}},
		{"KIND:lib", []string{"kit"}},
		{"lang:go", []string{"orders", "kit"}},
		{"changed", []string{"web"}},
		{"changed kind:lib", nil},
		{"kind:app go", []string{"orders"}},
		{"zzz", nil},
		{"web", []string{"web"}},
	} {
		listed, _ := l.WithFilter(tc.filter).Listed()
		var got []string
		for _, p := range listed {
			got = append(got, p.Name)
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("%q lists %v, want %v", tc.filter, got, tc.want)
		}
	}
}

func TestTheCursorStaysOnItsProjectOrMovesToTheFirstListedOne(t *testing.T) {
	l := sample().SelectProject("web")
	if got := l.WithFilter("kind:app").SelectedName(); got != "web" {
		t.Fatalf("web passes the filter and keeps the cursor, got %q", got)
	}
	if got := l.WithFilter("lang:go").SelectedName(); got != "orders" {
		t.Fatalf("web does not pass: the first listed project is selected, got %q", got)
	}
	if _, ok := l.WithFilter("zzz").Selected(); ok {
		t.Fatal("nothing is selected when the filter lists nothing")
	}
	if l.WithFilter("kind:app").Filter() != "kind:app" || l.Len() != 5 {
		t.Fatal("the filter is kept and the list still holds every project")
	}
}

func TestTheCursorMovesOnlyOverTheListedProjects(t *testing.T) {
	l := sample().WithFilter("lang:go") // orders, kit
	if l.SelectedName() != "orders" {
		t.Fatalf("starts on %q", l.SelectedName())
	}
	if l = l.Move(1); l.SelectedName() != "kit" {
		t.Fatalf("Down skips web and goes to kit, got %q", l.SelectedName())
	}
	if l = l.Move(1); l.SelectedName() != "kit" {
		t.Fatalf("Down stops at the last listed project, got %q", l.SelectedName())
	}
	if l = l.Home(); l.SelectedName() != "orders" {
		t.Fatalf("Home goes to the first listed project, got %q", l.SelectedName())
	}
	if l = l.End(); l.SelectedName() != "kit" {
		t.Fatalf("End goes to the last listed project, got %q", l.SelectedName())
	}
	_, indexes := l.Listed()
	if !slices.Equal(indexes, []int{0, 2}) {
		t.Fatalf("the indexes are those of Projects: %v", indexes)
	}
}
