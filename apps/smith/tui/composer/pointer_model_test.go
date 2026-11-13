package composer

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/muesli/termenv"
)

// cellOf is the screen cell where text is drawn on the screen of m.
func cellOf(t *testing.T, m Model, text string) (x, y int) {
	t.Helper()
	for row, line := range m.Lines(theme.Default()) {
		plain := pane.Strip(line)
		if i := strings.Index(plain, text); i >= 0 {
			return m.rect.X + len([]rune(plain[:i])), m.rect.Y + row
		}
	}
	t.Fatalf("%q is not on the screen:\n%s", text, pane.Strip(strings.Join(m.Lines(theme.Default()), "\n")))
	return 0, 0
}

func TestAClickOnARowTicksItAndASecondClickUnticksIt(t *testing.T) {
	m := screen()
	x, y := cellOf(t, m, "[ ] Python")
	m, _ = m.Click(x+1, y)
	if m.lang == nil || m.lang.Name != "python" {
		t.Fatalf("a click on the language row ticks it: %+v", m.lang)
	}
	if !strings.Contains(pane.Strip(text(m)), "[x] Python") {
		t.Fatalf("the box shows the tick:\n%s", pane.Strip(text(m)))
	}
	x, y = cellOf(t, m, "[x] Python")
	m, _ = m.Click(x+1, y)
	if m.lang != nil {
		t.Fatal("a second click unticks it")
	}
}

func TestAClickOnAChoiceRowTicksIt(t *testing.T) {
	m := screen()
	x, y := cellOf(t, m, "[ ] ESLint")
	m, _ = m.Click(x+1, y)
	if got := m.Ticked(); len(got) != 1 || got[0] != "eslint" {
		t.Fatalf("ticked %v", got)
	}
}

func TestAClickOnAHeadingOrOutsideTheListTicksNothing(t *testing.T) {
	m := screen()
	m, _ = m.Click(0, 0)
	m, _ = m.Click(m.rect.X+3, m.rect.Y+1) // the search box
	if len(m.Ticked()) != 0 || m.lang != nil {
		t.Fatalf("ticked %v %+v", m.Ticked(), m.lang)
	}
}

func TestTheRowUnderThePointerTakesTheRowHoverLook(t *testing.T) {
	defer lipgloss.SetColorProfile(lipgloss.ColorProfile())
	lipgloss.SetColorProfile(termenv.TrueColor)
	m := screen()
	x, y := cellOf(t, m, "[ ] ESLint")
	row := y - m.rect.Y
	rest := m.Lines(theme.Default())[row]
	hovered := m.Hover(x+1, y)
	if got := hovered.Lines(theme.Default())[row]; got == rest {
		t.Fatal("the row under the pointer must show the row hover look")
	}
	if away := hovered.Hover(0, 0); away.Lines(theme.Default())[row] != rest {
		t.Fatal("the row is at rest when the pointer leaves")
	}
}

func TestAClickOnAnAnswerOfTheOpenQuestionPicksIt(t *testing.T) {
	m := typed(screen(), "serverless")
	m, _ = press(m, "down", " ")
	if m.stage != stQuestion {
		t.Fatalf("Azure Functions asks for a trigger first, stage %v", m.stage)
	}
	x, y := cellOf(t, m, "Timer trigger")
	m, _ = m.Click(x+1, y)
	if got := m.Ticked(); len(got) != 2 || got[1] != "timer-trigger" {
		t.Fatalf("the click picks the answer: %v", got)
	}
}
