package palette

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

var ui = []Entry{
	{ID: "run.build", Section: "Build", Title: "Build web", Hint: "smith run build -p web", Shortcut: "b"},
	{ID: "run.stop", Section: "Build", Title: "Stop", Hint: "cancel the run", Shortcut: "Ctrl+C", Unavailable: "nothing is running"},
	{ID: "ws.doctor", Section: "Health", Title: "Doctor", Hint: "smith doctor", Shortcut: "d"},
}

func typed(p Palette, text string) Palette {
	for _, r := range text {
		p, _ = p.Key(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return p
}

func enter() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyEnter} }

func TestTheListIsGroupedBySectionWithShortcutsAndCommandHints(t *testing.T) {
	text := strings.Join(New(ui).Lines(theme.Default(), 70, 12), "\n")
	for _, want := range []string{"Build", "Build web", "smith run build -p web", "Health", "Doctor"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
	if strings.Contains(text, "Stop") {
		t.Fatalf("a command that cannot work now is hidden until searched:\n%s", text)
	}
}

func TestSearchingFindsByTitleByCommandLineAndBySectionAndShowsTheUnusableLast(t *testing.T) {
	if _, res := typed(New(ui), "doc").Key(enter()); res.CommandID != "ws.doctor" {
		t.Fatalf("%+v", res)
	}
	if _, res := typed(New(ui), "run build").Key(enter()); res.CommandID != "run.build" {
		t.Fatalf("the command line is searchable: %+v", res)
	}
	if _, res := typed(New(ui), "health").Key(enter()); res.CommandID != "ws.doctor" {
		t.Fatalf("the section is searchable: %+v", res)
	}
}

func TestPickingAShellCommandReportsItsID(t *testing.T) {
	_, res := typed(New(ui), "build").Key(enter())
	if res.CommandID != "run.build" || !res.Closed {
		t.Fatalf("%+v", res)
	}
}

func TestAnUnavailableShellCommandSaysWhyAndStaysOpen(t *testing.T) {
	_, res := typed(New(ui), "stop").Key(enter())
	if res.CommandID != "" || res.Closed || !strings.Contains(res.Refused, "nothing is running") {
		t.Fatalf("%+v", res)
	}
}

func TestNothingMatchingOffersNoCommandToConfirm(t *testing.T) {
	p := typed(New(ui), "lodash")
	if _, res := p.Key(enter()); res.CommandID != "" || res.Closed || res.Refused != "" {
		t.Fatalf("free text must not become a command: %+v", res)
	}
}

func TestEscClosesThePalette(t *testing.T) {
	if _, res := New(ui).Key(tea.KeyMsg{Type: tea.KeyEsc}); !res.Closed {
		t.Fatal("esc closes")
	}
}

func TestClickingARowChoosesItAndAHeadingIsNoRow(t *testing.T) {
	p := New(ui)
	if p.RowAt(0, 9) != -1 {
		t.Fatal("the first line is the Build heading")
	}
	i := p.RowAt(1, 9)
	if i != 0 {
		t.Fatalf("line 1 is match %d", i)
	}
	if _, res := p.Choose(i); res.CommandID != "run.build" {
		t.Fatalf("%+v", res)
	}
	if p.RowAt(50, 9) != -1 {
		t.Fatal("below the list")
	}
}

func TestTheWheelAndTheArrowsMoveTheCursor(t *testing.T) {
	p := New(ui).Scroll(1)
	if p.cursor != 1 {
		t.Fatalf("%d", p.cursor)
	}
	p, _ = p.Key(tea.KeyMsg{Type: tea.KeyUp})
	if p.cursor != 0 {
		t.Fatalf("%d", p.cursor)
	}
	if New(ui).Scroll(99).cursor != 1 || New(ui).Scroll(-9).cursor != 0 {
		t.Fatal("clamped")
	}
}

func TestEveryLineFitsTheWidthAtAnySize(t *testing.T) {
	for _, s := range [][2]int{{0, 0}, {1, 1}, {5, 2}, {20, 5}, {30, 9}, {74, 18}, {136, 20}} {
		p := New(ui)
		for _, lines := range [][]string{p.Lines(theme.Default(), s[0], s[1]), typed(p, "zzzz").Lines(theme.Default(), s[0], s[1])} {
			if s[0] < 1 || s[1] < 1 {
				continue
			}
			for _, l := range lines {
				if pane.Width(l) > s[0] {
					t.Fatalf("%v: %q is %d wide", s, l, pane.Width(l))
				}
			}
		}
	}
}
