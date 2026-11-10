package guide

import (
	"slices"
	"testing"
)

func TestTheLanguageFilterFindsALanguageByTheTitleAndAliasesOfItsChoices(t *testing.T) {
	offers, err := ParseChoices(`[
 {"axis":"framework","name":"electron","title":"Electron","aliases":["desktop","windows"],"description":"a desktop app","language":"node"},
 {"axis":"framework","name":"fastapi","title":"FastAPI","aliases":["api"],"description":"an API","language":"python"}
]`)
	if err != nil {
		t.Fatal(err)
	}
	options := withChoiceNames([]Option{{Label: "Node", Value: "node"}, {Label: "Python", Value: "python"}}, offers)
	for _, want := range []string{"electron", "Electron", "desktop", "windows"} {
		if !slices.Contains(options[0].Also, want) {
			t.Fatalf("Node is found by %q: %v", want, options[0].Also)
		}
	}
	if slices.Contains(options[1].Also, "desktop") {
		t.Fatalf("Python has no desktop shell: %v", options[1].Also)
	}
	if !MatchesFilter(options[0], "desktop") {
		t.Fatal("typing desktop in the language step finds Node")
	}
}
