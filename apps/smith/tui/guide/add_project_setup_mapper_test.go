package guide

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTheSetupQuestionOffersToMakeMyOwnRecipe(t *testing.T) {
	var own *Option
	for i, o := range addSetupStep().Options {
		if o.Wizard == WizardRecipe {
			own = &addSetupStep().Options[i]
		}
	}
	if own == nil || own.Label != "Make my own recipe…" {
		t.Fatalf("the setup question has no way to the recipe form: %+v", addSetupStep().Options)
	}
}

func TestAnOptionWithAWizardEndsTheFormAndNamesIt(t *testing.T) {
	m := NewModel([]Action{{
		Title: "Add a thing",
		Steps: []Step{{Key: "how", Prompt: "How?", Options: []Option{
			{Label: "Defaults", Value: "defaults"},
			{Label: "Own recipe", Value: "recipe", Wizard: WizardRecipe},
		}}, {Key: "name", Prompt: "Name?"}},
		Argv: func(map[string][]string) []string { return []string{"add"} },
	}})
	m, _ = m.Pick("Add a thing")
	m, _ = m.Handle(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Handle(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.Done() || m.Wizard != WizardRecipe || len(m.Argv) != 0 {
		t.Fatalf("done %v, wizard %q, argv %v", m.Done(), m.Wizard, m.Argv)
	}
}

func TestActionPresentingFindsAFormByItsCatalogueID(t *testing.T) {
	m := NewModel([]Action{{Title: "Why is this here?", Presents: "why"}})
	if a, ok := m.ActionPresenting("why"); !ok || a.Title != "Why is this here?" {
		t.Fatal("the why form was not found by its id")
	}
	if _, ok := m.ActionPresenting("boundaries"); ok {
		t.Fatal("found a form the menu does not have")
	}
}
