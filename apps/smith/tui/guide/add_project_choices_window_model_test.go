package guide

import (
	"os"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// choicesStepOfAWebsite walks the real catalogue of the node plugin to the
// choices step of a website whose framework is decided in the next question.
func choicesStepOfAWebsite(t *testing.T) Model {
	t.Helper()
	raw, err := os.ReadFile("testdata/node_choices.json")
	if err != nil {
		t.Fatal(err)
	}
	offers, err := ParseChoices(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	var add Action
	for _, a := range Actions([]string{"node"}, nil, nil) {
		if a.Title == "Add a project" {
			add = a
		}
	}
	m, ok := NewModel([]Action{AddProjectAction(add, offers, nil)}).Pick("Add a project")
	if !ok {
		t.Fatal("no form")
	}
	m = press(m, "enter")                     // A website
	m = press(m, "enter")                     // Decide in the next question
	m = press(m, "s", "i", "t", "e", "enter") // the name
	if !strings.Contains(m.View(), "Build it from choices?") {
		t.Fatalf("not on the choices step:\n%s", m.View())
	}
	return m
}

// rowsOfTheList counts the lines of the pane that belong to the list of options.
func rowsOfTheList(m Model, w, h int) (rows int, text string) {
	lines, at := m.Window(theme.Default(), w, h)
	for i := range lines {
		lines[i] = pane.Strip(lines[i])
	}
	return len(at), strings.Join(lines, "\n")
}

func TestTheChoicesStepGivesTheListTwelveRowsInThePaneOfASmallTerminal(t *testing.T) {
	// The pane is the terminal minus the tab bar, the border and the footer.
	for _, sc := range []struct {
		name string
		w, h int
		min  int
	}{{"110x32", 108, 25, 12}, {"80x24", 78, 17, 8}} {
		t.Run(sc.name, func(t *testing.T) {
			rows, text := rowsOfTheList(choicesStepOfAWebsite(t), sc.w, sc.h)
			if rows < sc.min {
				t.Fatalf("the list got %d rows:\n%s", rows, text)
			}
			if !strings.Contains(text, wordsHint) || strings.Contains(text, "shell = ") {
				t.Fatalf("the words must fold to one line:\n%s", text)
			}
			if strings.Contains(text, "[x]") {
				t.Fatalf("deciding later ticks nothing:\n%s", text)
			}
		})
	}
}

func TestTheGlossaryOpensOnTheChoicesStepWhenThePaneIsTall(t *testing.T) {
	rows, text := rowsOfTheList(choicesStepOfAWebsite(t), 108, 60)
	if !strings.Contains(text, "shell = ") || rows < 12 {
		t.Fatalf("a tall pane keeps the words (%d rows):\n%s", rows, text)
	}
}
