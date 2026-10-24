package settingsview

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func model() Model {
	return FromResult(Result{Settings: []Setting{
		{Field: configedit.Field{Path: "ci.providers", Type: "array", Enum: []string{"github", "azure"}, Description: `Pipelines to generate. Default: ["github"].`}},
		{Field: configedit.Field{Path: "cache.max_size_mb", Type: "integer", Description: "Prune past this size. Default: 2048."}, Value: "4096", Set: true},
	}})
}

func render(m Model, w, h int) string {
	lines, _ := m.View(theme.Default(), w, h)
	for _, l := range lines {
		if pane.Width(l) > w {
			panic("line wider than the pane: " + pane.Strip(l))
		}
	}
	return pane.Strip(strings.Join(lines, "\n"))
}

func TestAnUnsetSettingShowsItsDefaultValue(t *testing.T) {
	out := render(model(), 80, 24)
	if !strings.Contains(out, `default: ["github"]`) {
		t.Fatalf("the default value is missing:\n%s", out)
	}
}

func TestTheSelectedSettingIsExplainedUnderTheListEvenOnAShortPane(t *testing.T) {
	out := render(model(), 80, 14)
	for _, want := range []string{"Pipelines to generate.", "any of github, azure", "its default"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
}

func TestEditingHappensOnTheRowWithTheAllowedValuesShown(t *testing.T) {
	m, _ := model().Update(tea.KeyMsg{Type: tea.KeyEnter})
	out := render(m, 80, 24)
	if !strings.Contains(out, "ci.providers") || !strings.Contains(out, "[") || !strings.Contains(out, "accepts  any of github, azure") {
		t.Fatalf("the editor should sit on the row and list the allowed values:\n%s", out)
	}
}
