package cachewizard

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// cellOf is the screen cell where the dialog draws text.
func cellOf(t *testing.T, m Model, text string) (int, int) {
	t.Helper()
	for i, l := range m.Lines(theme.Default(), 0, time.Now()) {
		plain := pane.Strip(l)
		if at := strings.Index(plain, text); at >= 0 {
			return m.Rect().X + pane.Width(plain[:at]), m.Rect().Y + i
		}
	}
	t.Fatalf("%q is not drawn", text)
	return 0, 0
}

func TestOneClickPicksAnOption(t *testing.T) {
	m := listOf("azure", "s3", "https").Place(100, 30)
	x, y := cellOf(t, m, "2) s3")
	_, out := m.Click(theme.Default(), x+1, y)
	if out.Reply == nil || out.Reply.Value != "s3" {
		t.Fatalf("one click picks the option: %+v", out)
	}
}

func TestTheButtonsAreCentredAndLitUnderThePointer(t *testing.T) {
	m := listOf("azure", "s3").Place(100, 30)
	x, y := cellOf(t, m, "Cancel (Esc)")
	want := button.RowSpans(theme.Default(), []string{"Cancel"}, []string{"esc"}, m.Rect().W-4)[0][0]
	// The button's cell starts one column before its label.
	if got := x - 1 - m.Rect().X; got != want {
		t.Fatalf("the Cancel button starts at %d, want %d (centred)", got, want)
	}
	// The terminal of a test has no colour, so the hover is read from the model.
	if over := m.Hover(theme.Default(), x+1, y).overButton; over != [2]int{y - m.Rect().Y, want - 2} {
		t.Fatalf("the button under the pointer is not the hovered one: %v", over)
	}
	if over := m.Hover(theme.Default(), 0, 0).overButton; over != [2]int{} {
		t.Fatalf("a pointer off the dialog lights no button: %v", over)
	}
}
