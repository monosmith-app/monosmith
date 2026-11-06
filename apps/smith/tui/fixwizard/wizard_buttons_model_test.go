package fixwizard

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/muesli/termenv"
)

func init() {
	// Draw escape codes even though the tests have no terminal, so the button states can be told apart.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func failed(t *testing.T) Wizard {
	w, _ := press(t, New(safe(), show).Place(100, 30), "enter")
	return w.Fail(&doctorfix.StepError{Number: 1, Total: 1, Step: protocol.FixStep{Argv: []string{"npm", "install"}}, Err: errors.New("exit 1")})
}

func running(t *testing.T) Wizard {
	w, _ := press(t, New(safe(), show).Place(100, 30), "enter")
	return w
}

func TestAClickOnEveryButtonPressesItsKey(t *testing.T) {
	styled := Style{Name: "Adopt", Apply: "Go", Decline: "No", Skip: "Skip", Back: true}
	for name, w := range map[string]Wizard{
		"review":  New(safe(), show).Styled(styled).Place(100, 30),
		"asking":  New(question(), show).Place(100, 30),
		"failing": failed(t),
		"running": running(t),
	} {
		r := w.Rect()
		for i, b := range w.buttons() {
			_, byClick := w.Click(r.X+w.buttonSpans(r.W - 4)[i][0]+1, r.Y+r.H-2)
			_, byKey := w.Key(b.key)
			if byClick != byKey {
				t.Fatalf("%s: button %q: click %+v, key %q %+v", name, b.label, byClick, b.key, byKey)
			}
		}
	}
}

func TestTheButtonUnderThePointerIsDrawnInTheHoverState(t *testing.T) {
	w := New(safe(), show).Place(100, 30)
	r := w.Rect()
	rest := strings.Join(w.Lines(theme.Default(), 0), "\n")
	last := len(w.buttons()) - 1
	if hovered := strings.Join(w.Hover(r.X+w.buttonSpans(r.W - 4)[last][0]+1, r.Y+r.H-2).Lines(theme.Default(), 0), "\n"); hovered == rest {
		t.Fatal("hover does not change the button")
	}
}
