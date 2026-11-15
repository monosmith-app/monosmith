package guide

import (
	"reflect"
	"testing"
)

func swapOffers() []ChoiceOffer {
	return []ChoiceOffer{
		{Axis: "test", Name: "jest", Language: "node", Default: true},
		{Axis: "test", Name: "vitest", Language: "node"},
		{Axis: "lint", Name: "eslint", Language: "node"},
		{Axis: "test", Name: "gotest", Language: "go"},
	}
}

func TestSwapAsksOnlyTheDecisionsOfTheLanguageOfTheProject(t *testing.T) {
	base := Action{Title: "Swap a choice", Steps: []Step{{Key: "project"}}, Argv: SwapArgv}
	a := SwapProjectAction(base, swapOffers(), []SwapProject{{"web", "node"}, {"core", "go"}})
	if a.Steps[0].Key != "project" || len(a.Steps[0].Options) != 2 {
		t.Fatalf("the project is picked from the list: %+v", a.Steps[0])
	}
	skipped := func(project string) map[string]bool {
		answers := map[string][]string{"project": {project}, "axis:node": {"test"}, "axis:go": {"test"}}
		out := map[string]bool{}
		for _, s := range a.Steps[1:] {
			out[s.Key] = s.Skip(answers)
		}
		return out
	}
	node := skipped("web")
	if node["axis:node"] || node["choice:node:test"] || !node["choice:node:lint"] || !node["axis:go"] || !node["choice:go:test"] {
		t.Fatalf("a node project is asked the node decision and the chosen axis: %v", node)
	}
	if got := skipped("core"); got["axis:go"] || got["choice:go:test"] || !got["axis:node"] {
		t.Fatalf("a go project is asked the go decision: %v", got)
	}
}

func TestSwapBuildsTheCommandFromTheAnswersOfOneLanguage(t *testing.T) {
	want := []string{"swap", "web", "test", "vitest"}
	for name, answers := range map[string]map[string][]string{
		"plain form":     {"project": {"web"}, "axis": {"test"}, "choice": {"vitest"}},
		"pick lists":     {"project": {"web"}, "axis:node": {"test"}, "choice:node:test": {"vitest"}},
		"an empty other": {"project": {"web"}, "axis:node": {"test"}, "choice:node:test": {"vitest"}, "axis:go": nil},
	} {
		if got := SwapArgv(answers); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %v", name, got)
		}
	}
}

func TestSwapKeepsThePlainFormWhenThereIsNothingToPickFrom(t *testing.T) {
	base := Action{Title: "Swap a choice", Steps: []Step{{Key: "project"}, {Key: "axis"}, {Key: "choice"}}}
	if got := SwapProjectAction(base, nil, []SwapProject{{"web", "node"}}); len(got.Steps) != 3 {
		t.Fatal("no choices: the plain form stays")
	}
}
