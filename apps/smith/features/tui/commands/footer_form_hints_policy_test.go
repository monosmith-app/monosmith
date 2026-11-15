package commands

import "testing"

// TestAFormFooterDoesNotPromiseAKeyThatTypesIntoTheField keeps "n new workspace"
// off the footer while the name of a workspace is typed: n is a letter there.
func TestAFormFooterDoesNotPromiseAKeyThatTypesIntoTheField(t *testing.T) {
	has := func(s State, glyph string) bool {
		view, _ := Hints(s)
		for _, h := range view {
			if h.Glyph == glyph {
				return true
			}
		}
		return false
	}
	outside := State{View: TabGuide}
	if !has(outside, "n") {
		t.Fatal("the start screen does not offer n for a new workspace")
	}
	outside.Form = true
	if has(outside, "n") {
		t.Fatal("a form footer still offers n while a name is typed")
	}
}
