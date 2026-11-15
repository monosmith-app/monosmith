package projectlist

import "testing"

func sample() List {
	return New([]Project{
		{Name: "web", Kind: "app", Language: "node", Tasks: []Task{{Name: "build"}, {Name: "test"}}, Deps: []string{"kit"}, Affected: true},
		{Name: "sdk", Kind: "package", Language: "python"},
		{Name: "orders", Kind: "app", Language: "go", Tasks: []Task{{Name: "build"}}},
		{Name: "kit", Kind: "lib", Language: "go", Tasks: []Task{{Name: "test"}}, Dependents: []string{"web"}},
		{Name: "zeta", Kind: "tool"},
	})
}

func names(l List) []string {
	var out []string
	for _, p := range l.Projects() {
		out = append(out, p.Name)
	}
	return out
}

func TestProjectsAreOrderedByKindThenName(t *testing.T) {
	got := names(sample())
	want := []string{"orders", "web", "kit", "sdk", "zeta"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestTheCursorMovesWithinTheListAndStops(t *testing.T) {
	l := sample()
	if l.SelectedName() != "orders" {
		t.Fatalf("starts on %q", l.SelectedName())
	}
	if l = l.Move(-1); l.Cursor() != 0 {
		t.Fatal("moved above the first project")
	}
	if l = l.End(); l.SelectedName() != "zeta" || l.Move(1).Cursor() != 4 {
		t.Fatalf("end: %q", l.SelectedName())
	}
	if l.Home().SelectedName() != "orders" {
		t.Fatal("home")
	}
	if got := l.Home().Page(1, 3).Cursor(); got != 2 {
		t.Fatalf("a page of 3 rows moves 2, moved to %d", got)
	}
}

func TestSelectingByNameAndByIndex(t *testing.T) {
	l := sample().SelectProject("kit")
	if l.SelectedName() != "kit" {
		t.Fatal(l.SelectedName())
	}
	if l.Select(99).SelectedName() != "kit" {
		t.Fatal("an index out of range must change nothing")
	}
	if p, ok := l.Project("web"); !ok || !p.Affected {
		t.Fatal("web should be found by name")
	}
}

func TestAnEmptyListSelectsNothing(t *testing.T) {
	l := New(nil)
	if _, ok := l.Selected(); ok || l.Move(3).Cursor() != 0 {
		t.Fatal("an empty list has no selection")
	}
}
