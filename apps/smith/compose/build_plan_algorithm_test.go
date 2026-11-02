package compose

import (
	"strings"
	"testing"
)

func completeReact() Check { return CheckPicks(fakeOffers(), "node", []string{"react"}) }

func TestThePlanIsTheExactAddLineWithTheDefaultsWrittenOut(t *testing.T) {
	p := BuildPlan(PlanInput{Kind: "app", Name: "web", Check: completeReact(), Language: "node", Scaffold: "vite", Toolchain: []string{"vite", "eslint", "vitest"}})
	if p.Command != "smith add app web --lang node --with react,eslint,vitest" {
		t.Fatalf("command %q", p.Command)
	}
	if len(p.Notes) != 1 || !strings.Contains(p.Notes[0], "eslint and vitest are added by default") {
		t.Fatalf("a default is never silent: %v", p.Notes)
	}
}

func TestThePlanShowsANamePlaceholderAndTheAppKindByDefault(t *testing.T) {
	p := BuildPlan(PlanInput{Check: completeReact()})
	if !strings.HasPrefix(p.Command, "smith add app <name> ") {
		t.Fatalf("command %q", p.Command)
	}
}

func TestThePlanListsTheToolsOfTheSelectionOnceEach(t *testing.T) {
	source := func(choice, _ string) []Tool {
		if choice == "react" || choice == "eslint" {
			return []Tool{{Choice: choice, Name: "Node", Binary: "node", Install: "install Node"}}
		}
		return nil
	}
	p := BuildPlan(PlanInput{Name: "web", Check: completeReact(), Tools: source})
	if len(p.Tools) != 1 || p.Tools[0].Binary != "node" {
		t.Fatalf("the same tool from two picks is listed once: %+v", p.Tools)
	}
	if none := BuildPlan(PlanInput{Name: "web", Check: completeReact()}); len(none.Tools) != 0 {
		t.Fatalf("without a tool source there are no tools: %+v", none.Tools)
	}
}

func TestTheRecipeKeepsThePluginsAndSaysWhatItLoses(t *testing.T) {
	p := BuildPlan(PlanInput{Name: "web", Check: completeReact(), Language: "node", Scaffold: "vite", Toolchain: []string{"eslint"}})
	want := "smith recipe new --name react-eslint-vitest --fixed language=node --fixed scaffold=vite --fixed toolchain=eslint"
	if p.RecipeCommand != want {
		t.Fatalf("recipe %q", p.RecipeCommand)
	}
	if !strings.Contains(p.RecipeNote, "plugins, not the choices") {
		t.Fatalf("note %q", p.RecipeNote)
	}
}

func TestAnIncompleteSelectionHasNoCommandAndSaysWhy(t *testing.T) {
	p := BuildPlan(PlanInput{Check: CheckPicks(fakeOffers(), "node", []string{"react", "electron"})})
	if p.Command != "" || len(p.Notes) != 1 || !strings.Contains(p.Notes[0], "keep one") {
		t.Fatalf("%+v", p)
	}
}
