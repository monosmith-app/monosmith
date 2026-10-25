package commands

import "testing"

func TestAdoptIsHiddenInAnEmptyFolder(t *testing.T) {
	for _, id := range []string{"file.adopt"} {
		c, ok := ByID(id)
		if !ok {
			t.Fatalf("%s is gone", id)
		}
		if c.Visible(State{EmptyFolder: true}) {
			t.Fatalf("%s has nothing to work on in an empty folder", id)
		}
		if !c.Visible(State{}) {
			t.Fatalf("%s should show in a folder that holds something", id)
		}
	}
	if c, _ := ByID("file.new"); !c.Visible(State{EmptyFolder: true}) {
		t.Fatal("an empty folder can still start a workspace")
	}
}
