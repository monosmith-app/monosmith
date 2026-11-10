package releasepick

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/muesli/termenv"
)

func init() {
	// Draw escape codes even though the tests have no terminal, so the button states can be told apart.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func TestAClickOnEveryButtonDoesWhatItsKeyDoes(t *testing.T) {
	m := New(samplePlan()).Place(100, 30)
	r := m.Rect()
	for i, key := range buttonKeys {
		_, byClick := m.Click(r.X+m.buttons(r.W - 4)[i][0]+1, r.Y+r.H-2)
		_, byKey := m.Key(key)
		if byClick != byKey {
			t.Fatalf("button %d: click %+v, key %q %+v", i, byClick, key, byKey)
		}
	}
}

func TestTheButtonUnderThePointerIsDrawnInTheHoverState(t *testing.T) {
	m := New(samplePlan()).Place(100, 30)
	r := m.Rect()
	rest := strings.Join(m.Lines(theme.Default()), "\n")
	if hovered := strings.Join(m.Hover(r.X+m.buttons(r.W - 4)[1][0]+1, r.Y+r.H-2).Lines(theme.Default()), "\n"); hovered == rest {
		t.Fatal("hover does not change the button")
	}
}
