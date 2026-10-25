package detail

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestThePanelListsRowsAndClickableActions(t *testing.T) {
	p := Panel{
		Sections: []Section{{Heading: "web", Rows: []Row{{"kind", "app"}, {"language", "node"}}, Items: nil}, {Heading: "Depends on", Items: []string{"kit"}}},
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
	p := Panel{Sections: []Section{{Rows: []Row{{"root", "apps/markobot-vscode-extension/src"}}}}}
	lines, _ := p.View(theme.Default(), 26, "")
	if len(lines) != 1 || !strings.Contains(lines[0], "…") || !strings.Contains(lines[0], "apps/") || !strings.Contains(lines[0], "/src") {
		t.Fatalf("the path should keep both ends on one line:\n%s", strings.Join(lines, "\n"))
	}
}

func TestAValueOfSeveralWordsWrapsAtTheSpaces(t *testing.T) {
	p := Panel{Sections: []Section{{Rows: []Row{{"about", "builds the web app and its worker for the cloud"}}}}}
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
		Sections: []Section{{Heading: heading, Rows: []Row{{"state", "warning"}}}},
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
