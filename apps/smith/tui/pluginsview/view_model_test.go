package pluginsview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func sampleModel() Model {
	return Model{Loaded: true, Rows: []Row{
		{Name: "node", Builtin: true, Switchable: true, Enabled: true, Roles: []string{"language", "toolchain", "scaffold", "sca-extra"}, UsedBy: []string{"alpha-service", "beta-service", "gamma-service"}},
		{Name: "bicep", Builtin: true, Switchable: true, Enabled: true, Roles: []string{"language", "toolchain", "scaffold", "sca-extra"}},
	}}
}

func TestViewKeepsEveryColumnInsideItsWidth(t *testing.T) {
	for _, w := range []int{50, 70, 100} {
		lines, _ := sampleModel().View(theme.Default(), w, 30)
		for _, l := range lines {
			if got := pane.Width(l); got > w {
				t.Fatalf("w=%d: line is %d cells: %q", w, got, pane.Strip(l))
			}
		}
	}
}

func TestViewMarksCutTextWithAnEllipsis(t *testing.T) {
	lines, _ := sampleModel().View(theme.Default(), 70, 30)
	text := pane.Strip(strings.Join(lines, "\n"))
	if !strings.Contains(text, "…") {
		t.Fatalf("a cut roles cell should end in an ellipsis:\n%s", text)
	}
}

func TestDetailPhrasesTheToggleAsASentence(t *testing.T) {
	lines, _ := sampleModel().View(theme.Default(), 100, 30)
	text := pane.Strip(strings.Join(lines, "\n"))
	if !strings.Contains(text, "Press E to disable, or run: smith plugin disable node") {
		t.Fatalf("detail should read as a sentence:\n%s", text)
	}
}
