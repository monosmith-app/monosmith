package inspector

import (
	"strings"
	"testing"

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

func TestAnEmptyPanelSaysSo(t *testing.T) {
	lines, _ := Panel{}.View(theme.Default(), 30, "")
	if !strings.Contains(lines[0], "Nothing selected") {
		t.Fatal(lines)
	}
}
