package dialog

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

// buttonCell is a cell inside button i of the dialog's button row.
func buttonCell(c Confirm, i int) (x, y int) {
	r := c.Rect()
	return r.X + c.buttonsRow(r.W - 4)[i][0] + 1, r.Y + r.H - 2
}

func TestAClickOnEveryButtonDoesWhatItsKeyDoes(t *testing.T) {
	ask := NewConfirm("Release", "Why", []string{"release", "--yes"}).Place(100, 30)
	for i, key := range []string{"enter", "esc"} {
		x, y := buttonCell(ask, i)
		_, byClick := ask.Click(x, y)
		_, byKey := ask.Key(key)
		if byClick != byKey {
			t.Fatalf("button %d: click %+v, key %q %+v", i, byClick, key, byKey)
		}
	}
	note := NewNotice("Oops", "It broke", nil).Place(100, 30)
	x, y := buttonCell(note, 0)
	if _, out := note.Click(x, y); !out.Done || out.Confirmed {
		t.Fatalf("the OK button closes and confirms nothing: %+v", out)
	}
}

func TestTheButtonUnderThePointerIsDrawnInTheHoverState(t *testing.T) {
	c := NewConfirm("Release", "Why", []string{"release", "--yes"}).Place(100, 30)
	c.focus = 1 // Cancel is the focused one, so Run is not drawn active.
	rest := strings.Join(c.Lines(theme.Default()), "\n")
	x, y := buttonCell(c, 0)
	if hovered := strings.Join(c.Hover(x, y).Lines(theme.Default()), "\n"); hovered == rest {
		t.Fatal("hover does not change the button")
	}
}
