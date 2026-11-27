package projectpage

import "testing"

func TestSectionsGoRoundAndTheirIDsNameThem(t *testing.T) {
	if Tasks.Step(-1) != Settings || Settings.Step(1) != Tasks || Tasks.Step(2) != Runs {
		t.Fatal("the tabs go round")
	}
	for _, s := range Sections {
		if got, ok := SectionOf(s.ID()); !ok || got != s {
			t.Fatalf("%v: SectionOf(%q) = %v, %v", s, s.ID(), got, ok)
		}
	}
	if _, ok := SectionOf("x:run-menu"); ok {
		t.Fatal("an ID of another button names no tab")
	}
}
