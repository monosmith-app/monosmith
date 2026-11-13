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

var intents = []Entry{
	{ID: "deps.update", Section: "Dependencies", Title: "Update the marked ones", Hint: "smith deps up", Keywords: "update, outdated, upgrade packages, newer versions"},
	{ID: "ws.undo", Section: "Health", Title: "Undo the last change…", Hint: "smith undo", Keywords: "revert, rollback, go back"},
	{ID: "ws.doctor", Section: "Health", Title: "Check my project (Doctor)", Hint: "smith doctor", Keywords: "broken, problems, health, doctor"},
	{ID: "ws.cache-clear", Section: "Health", Title: "Speed-up storage (Cache): delete it all…", Hint: "smith cache clear", Keywords: "speed, slow, cache"},
}

func TestAnIntentWordFindsTheCommandThatHoldsItAsAKeyword(t *testing.T) {
	for typedText, want := range map[string]string{"outdated": "deps.update", "upgrade packages": "deps.update", "rollback": "ws.undo", "broken": "ws.doctor", "slow": "ws.cache-clear", "doctor": "ws.doctor"} {
		if _, res := typed(New(intents), typedText).Key(enter()); res.CommandID != want {
			t.Errorf("%q chose %q, want %q", typedText, res.CommandID, want)
		}
	}
}

func TestKeywordsAreSearchableButNeverShown(t *testing.T) {
	text := strings.Join(New(intents).Lines(theme.Default(), 90, 14), "\n")
	if strings.Contains(text, "rollback") || strings.Contains(text, "outdated") {
		t.Fatalf("keywords must stay hidden:\n%s", text)
	}
}

func TestNothingFoundShowsTheClosestThreeCommandsToChooseFrom(t *testing.T) {
	p := typed(New(intents), "doctr")
	text := strings.Join(p.Lines(theme.Default(), 90, 14), "\n")
	if !strings.Contains(text, "Did you mean") || strings.Contains(text, "no command matches") {
		t.Fatalf("a close typo offers commands:\n%s", text)
	}
	if got := p.matches(); len(got) == 0 || len(got) > 3 {
		t.Fatalf("suggestions: %d", len(got))
	}
	if len(p.matches()) > 1 {
		var res Result
		if p, res = p.Key(enter()); res.CommandID != "" || res.Closed || res.Refused == "" {
			t.Fatalf("with several guesses the first Enter only asks: %+v", res)
		}
	}
	if _, res := p.Key(enter()); res.CommandID != "ws.doctor" {
		t.Fatalf("the closest is the first: %+v", res)
	}
}

func TestTypingAfterTheGuessQuestionAsksAgain(t *testing.T) {
	p := typed(New(intents), "doctr")
	if len(p.matches()) < 2 {
		t.Skip("one guess runs at once")
	}
	p, _ = p.Key(enter())
	p, _ = p.Key(tea.KeyMsg{Type: tea.KeyDown})
	if _, res := p.Key(enter()); res.CommandID != "" {
		t.Fatalf("moving to another guess asks again: %+v", res)
	}
}

func TestNothingCloseStillSaysNoCommandMatches(t *testing.T) {
	text := strings.Join(typed(New(intents), "zzqxj").Lines(theme.Default(), 90, 14), "\n")
	if !strings.Contains(text, "no command matches") {
		t.Fatalf("%s", text)
	}
}

func TestTheInputHasOnePointerAndACutRowHasOneEllipsis(t *testing.T) {
	long := []Entry{{ID: "a", Section: "S", Title: "Name", Hint: strings.Repeat("a long hint ", 12), Shortcut: "x"}}
	lines := New(long).Lines(theme.Default(), 50, 6)
	if got := pane.Strip(lines[0]); strings.Count(got, ">")+strings.Count(got, "›") != 1 {
		t.Fatalf("one pointer in the input line: %q", got)
	}
	for _, l := range lines {
		if strings.Contains(l, "……") {
			t.Fatalf("doubled ellipsis: %q", l)
		}
	}
}

func TestTheSelectedCommandHasTheCursorMarkOfEveryPickListAndOnlyIt(t *testing.T) {
	lines := New(ui).Lines(theme.Default(), 70, 12)
	marked := 0
	for _, l := range lines[1:] {
		plain := pane.Strip(l)
		if strings.HasPrefix(plain, "› ") {
			marked++
			if !strings.Contains(plain, "Build web") {
				t.Fatalf("the mark is on another row: %q", plain)
			}
		}
		if pane.Width(l) > 70 {
			t.Fatalf("row wider than the palette: %q", plain)
		}
	}
	if marked != 1 {
		t.Fatalf("%d rows carry the cursor mark, want 1:\n%s", marked, strings.Join(lines, "\n"))
	}
}
