package packageprompt

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/muesli/termenv"
)

func init() {
	// Draw escape codes even though the tests have no terminal, so the button states can be told apart.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func TestAClickOnEveryButtonDoesWhatItsKeyDoes(t *testing.T) {
	for _, p := range []Prompt{typeText(NewInstall("web", []string{"web"}).Place(100, 30), "zod"), NewRemove("web", []string{"web"}).Place(100, 30)} {
		r := p.rect
		keys := []tea.KeyMsg{{Type: tea.KeyEnter}, {Type: tea.KeyEsc}}
		for i, k := range keys {
			p.focus = focusPackages
			_, byClick := p.Click(r.X+p.buttonsRow(r.W - 4)[i][0]+1, r.Y+r.H-2)
			_, byKey := p.Key(k)
			if byClick != byKey {
				t.Fatalf("button %d: click %+v, key %q %+v", i, byClick, k.String(), byKey)
			}
		}
	}
}

func TestTheButtonUnderThePointerIsDrawnInTheHoverState(t *testing.T) {
	p := NewInstall("web", []string{"web"}).Place(100, 30)
	r := p.rect
	rest := strings.Join(p.Lines(theme.Default()), "\n")
	if hovered := strings.Join(p.Hover(r.X+p.buttonsRow(r.W - 4)[1][0]+1, r.Y+r.H-2).Lines(theme.Default()), "\n"); hovered == rest {
		t.Fatal("hover does not change the button")
	}
}
