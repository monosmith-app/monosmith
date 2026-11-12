package detail

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/muesli/termenv"
)

func TestThePanelListsRowsAndClickableActions(t *testing.T) {
	p := Panel{
		Sections: []Section{{Heading: "web", Rows: []Row{{Label: "kind", Value: "app"}, {Label: "language", Value: "node"}}, Items: nil}, {Heading: "Depends on", Items: []string{"kit"}}},
		Actions:  []Action{{ID: "run.build", Label: "Run build", Shortcut: "Ctrl+B"}, {ID: "stop", Label: "Stop", Disabled: true}},
	}
	lines, hits := p.View(theme.Default(), 30, "")
	text := strings.Join(lines, "\n")
	for _, want := range []string{"web", "kind", "app", "language", "node", "Depends on", "kit", "Actions", "Run build", "Ctrl+B"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
	if len(hits) != 2 {
		t.Fatalf("hits %v", hits)
	}
	for row, id := range hits {
		if id == "run.build" && !strings.Contains(lines[row], "Run build") {
			t.Fatalf("row %d is %q", row, lines[row])
		}
	}
}

func TestALongPathIsShortenedInTheMiddleNotCutIntoFragments(t *testing.T) {
	p := Panel{Sections: []Section{{Rows: []Row{{Label: "root", Value: "apps/markobot-vscode-extension/src"}}}}}
	lines, _ := p.View(theme.Default(), 26, "")
	if len(lines) != 1 || !strings.Contains(lines[0], "…") || !strings.Contains(lines[0], "apps/") || !strings.Contains(lines[0], "/src") {
		t.Fatalf("the path should keep both ends on one line:\n%s", strings.Join(lines, "\n"))
	}
}

func TestAValueOfSeveralWordsWrapsAtTheSpaces(t *testing.T) {
	p := Panel{Sections: []Section{{Rows: []Row{{Label: "about", Value: "builds the web app and its worker for the cloud"}}}}}
	lines, _ := p.View(theme.Default(), 26, "")
	if len(lines) < 2 {
		t.Fatalf("expected a wrap:\n%s", strings.Join(lines, "\n"))
	}
	for _, l := range lines {
		for _, word := range strings.Fields(l) {
			if !strings.Contains("builds the web app and its worker for the cloud about", word) {
				t.Fatalf("a word was cut: %q", word)
			}
		}
	}
}

func TestAnEmptyPanelSaysSo(t *testing.T) {
	lines, _ := Panel{}.View(theme.Default(), 30, "")
	if !strings.Contains(lines[0], "Nothing selected") {
		t.Fatal(lines)
	}
}

func TestALongHeadingAndALongActionWrapInsteadOfBeingCut(t *testing.T) {
	const heading = "the scripts of the root package.json call files and tools that exist"
	const label = "Fix: install the tools the scripts start, or remove those scripts"
	p := Panel{
		Sections: []Section{{Heading: heading, Rows: []Row{{Label: "state", Value: "warning"}}}},
		Actions:  []Action{{ID: "fix", Label: label}, {ID: "again", Label: "Check again", Shortcut: "r"}},
	}
	const w = 40
	lines, hits := p.View(theme.Default(), w, "")
	text := strings.Join(lines, " ")
	for _, word := range strings.Fields(heading + " " + label) {
		if !strings.Contains(text, word) {
			t.Fatalf("the word %q was cut off:\n%s", word, strings.Join(lines, "\n"))
		}
	}
	for _, l := range lines {
		if got := pane.Width(l); got > w {
			t.Fatalf("line %q is %d cells, wider than %d", l, got, w)
		}
	}
	var fixRows int
	for _, id := range hits {
		if id == "fix" {
			fixRows++
		}
	}
	if fixRows < 2 {
		t.Fatalf("a wrapped action must be clickable on every line, got %d rows: %v", fixRows, hits)
	}
}

func TestAnActionExplainsItselfUnderItsLabelAndTheExplanationClicksLikeIt(t *testing.T) {
	p := Panel{
		ActionsTitle: "Run on web",
		Actions:      []Action{{ID: "x:run:build", Label: "Run build", Hint: "runs tsc -p tsconfig.json · after generate · result remembered"}},
	}
	lines, hits := p.View(theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII}), 40, "")
	text := pane.Strip(strings.Join(lines, "\n"))
	for _, want := range []string{"Run on web", "Run build", "runs tsc -p tsconfig.json", "result remembered"} {
		if !strings.Contains(text, want) {
			t.Fatalf("the panel lacks %q:\n%s", want, text)
		}
	}
	rows := 0
	for _, id := range hits {
		if id == "x:run:build" {
			rows++
		}
	}
	if rows < 2 {
		t.Fatalf("the action and its explanation click as one: %d rows hit\n%s", rows, text)
	}
}

func TestThePanelSplitsOnlyWhenItHasSeveralActionsAndADescription(t *testing.T) {
	for _, tc := range []struct {
		name string
		p    Panel
		want bool
	}{
		{"one action with a description", Panel{Actions: []Action{{ID: "a", Hint: "does a"}}}, false},
		{"two actions with no description", Panel{Actions: []Action{{ID: "a"}, {ID: "b"}}}, false},
		{"two actions, one described", Panel{Actions: []Action{{ID: "a"}, {ID: "b", Hint: "does b"}}}, true},
		{"one action and one more, described", Panel{Actions: []Action{{ID: "a", Hint: "does a"}}, More: []Action{{ID: "m"}}}, true},
	} {
		if got := tc.p.Splits(); got != tc.want {
			t.Errorf("%s: Splits() = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestTheChosenActionIsTheFirstUntilAnotherIsChosen(t *testing.T) {
	p := Panel{Actions: []Action{{ID: "a"}, {ID: "b"}}, More: []Action{{ID: "m"}}}
	for chosen, want := range map[string]string{"": "a", "gone": "a", "b": "b", "m": "m"} {
		p.Chosen = chosen
		if got, ok := p.ChosenAction(); !ok || got.ID != want {
			t.Errorf("Chosen %q gives %q, want %q", chosen, got.ID, want)
		}
	}
	if _, ok := (Panel{}).ChosenAction(); ok {
		t.Error("a panel with no actions has no chosen action")
	}
}

func TestASplitPanelShowsCompactChoosersThenTheChosenDescriptionAndOneRunButton(t *testing.T) {
	p := Panel{
		Actions: []Action{{ID: "build", Label: "Build", Shortcut: "b", Hint: "compiles the code"}, {ID: "test", Label: "Test", Shortcut: "t", Hint: "runs the tests"}},
		After:   []Section{{Heading: "Notes", Items: []string{"after the actions"}}},
		Chosen:  "test",
	}
	lines, hits := p.View(theme.Default(), 40, "")
	text := pane.Strip(strings.Join(lines, "\n"))
	if strings.Contains(text, "compiles the code") {
		t.Fatalf("only the chosen action is described:\n%s", text)
	}
	if strings.Count(text, "runs the tests") != 1 || strings.Count(text, "Test (t)") != 2 {
		t.Fatalf("the chosen action is a chooser and the run button, described once:\n%s", text)
	}
	build, test, run, desc, notes := -1, -1, -1, -1, -1
	for i, l := range lines {
		s := pane.Strip(l)
		switch {
		case strings.Contains(s, "Build (b)"):
			build = i
		case strings.Contains(s, "Test (t)") && test < 0:
			test = i
		case strings.Contains(s, "Test (t)"):
			run = i
		case strings.Contains(s, "runs the tests"):
			desc = i
		case strings.Contains(s, "Notes"):
			notes = i
		}
	}
	if build >= test || test+1 >= desc || desc >= run || run >= notes {
		t.Fatalf("order: choosers, rule, description, run button, then the sections after:\n%s", text)
	}
	if hits[build] != "build" || hits[test] != "test" || hits[run] != RunID {
		t.Fatalf("each chooser clicks its action and the run button clicks RunID: %v", hits)
	}
	if _, ok := hits[desc]; ok {
		t.Fatal("the description is not a button")
	}
}

func TestTheRunButtonLooksDisabledWhenTheChosenActionCannotRun(t *testing.T) {
	defer lipgloss.SetColorProfile(lipgloss.ColorProfile())
	lipgloss.SetColorProfile(termenv.TrueColor)
	th := theme.Default()
	p := Panel{Actions: []Action{{ID: "a", Label: "Alpha", Hint: "does a"}, {ID: "b", Label: "Beta", Hint: "does b"}}}
	on, rows := p.View(th, 40, "")
	p.Actions[0].Disabled = true
	off, _ := p.View(th, 40, "")
	run := -1
	for row, id := range rows {
		if id == RunID {
			run = row
		}
	}
	if run < 0 || on[run] == off[run] {
		t.Fatalf("the run button of an action that cannot run looks disabled: %q", off)
	}
	if on[1] != off[1] {
		t.Fatal("the chosen chooser keeps its selected look even when its action cannot run")
	}
}
