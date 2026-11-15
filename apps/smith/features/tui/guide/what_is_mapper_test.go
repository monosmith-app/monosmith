package guide

import (
	"os"
	"strings"
	"testing"
)

func TestEveryWordHasOneLineThatSaysWhatItIs(t *testing.T) {
	for _, term := range []string{"stack", "slice", "plugin", "recipe", "framework", "shell", "host", "test", "lint", "style"} {
		line := WhatIs(term)
		if line == "" || strings.Contains(line, "\n") || len(line) > 160 {
			t.Errorf("%s: want one short line, got %q", term, line)
		}
	}
	if WhatIs("klingon") != "" {
		t.Error("an unknown word has no line")
	}
}

func TestTheGlossaryOfTheDocsKnowsTheWordsTheFormsExplain(t *testing.T) {
	raw, err := os.ReadFile("../../../../docs/reference/glossary.md")
	if err != nil {
		t.Skip("the docs are not next to the code: ", err)
	}
	for _, term := range []string{"stack", "slice", "plugin", "recipe"} {
		if !strings.Contains(string(raw), "\n## "+term+"\n") {
			t.Errorf("docs/reference/glossary.md has no entry for %q: the form explains a word the docs do not", term)
		}
	}
}

func TestTheChoicesQuestionExplainsTheWordsOnItsList(t *testing.T) {
	offers, _ := ParseChoices(outcomeOffersJSON)
	note := choiceWords(offers)
	for _, want := range []string{"stack = the set of tools", "framework = how the app or library is written", "shell = what wraps the app", "host = where the code runs in the cloud"} {
		if !strings.Contains(note, want) {
			t.Errorf("the choices question does not explain %q:\n%s", want, note)
		}
	}
}

func TestTheFormsShowTheLineWhereTheWordFirstAppears(t *testing.T) {
	var plugins Action
	for _, a := range Actions(nil, []string{"node"}, nil) {
		if a.Presents == "plugin set" {
			plugins = a
		}
	}
	m, _ := NewModel([]Action{plugins}).Pick("Manage plugins")
	if view := m.View(); !strings.Contains(view, "plugin = a program that teaches smith something") {
		t.Errorf("Manage plugins does not say what a plugin is:\n%s", view)
	}
	recipe, _ := NewModel([]Action{{Title: "Write a recipe", Steps: NewRecipeSteps(), Argv: NewRecipeArgv}}).Pick("Write a recipe")
	if view := recipe.View(); !strings.Contains(view, "recipe = a file that says how every new project is made") {
		t.Errorf("Write a recipe does not say what a recipe is:\n%s", view)
	}
	for _, s := range NewWorkspaceSteps() {
		if s.Key == "architecture" && !strings.Contains(s.Note, "A slice is a folder that holds one outcome") {
			t.Errorf("the code layout question does not say what a slice is: %q", s.Note)
		}
	}
	add := outcomeForm(t)
	add = press(add, "d", "e", "s", "k", "t", "enter")
	if view := add.View(); !strings.Contains(view, "shell = what wraps the app as a program you install on a computer") {
		t.Errorf("the desktop question does not say what a shell is:\n%s", view)
	}
}
