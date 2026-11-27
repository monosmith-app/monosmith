package projectpage

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/muesli/termenv"
)

func headInput() HeadInput {
	return HeadInput{
		Path: "apps/orders", Kind: "app", Language: "go", Changed: true,
		Buttons: []Button{{ID: "run", Label: "Run…", Key: "R", Primary: true}, {ID: "open", Label: "Open"}},
		Section: Runs,
	}
}

func TestTheHeadSaysPathKindAndLanguageWithTheChangedMarkAndTheButtonsAtTheRight(t *testing.T) {
	lines, spans := Head(plain, headInput(), 80)
	first := pane.Strip(lines[0])
	if !strings.HasPrefix(first, " apps/orders · app · go  ● changed") || !strings.HasSuffix(strings.TrimRight(first, " "), "[Run… (R)]  [Open]") {
		t.Fatalf("the meta line is %q", first)
	}
	if len(spans[0]) != 2 || spans[0][0].ID != "run" || spans[0][1].ID != "open" {
		t.Fatalf("the buttons answer clicks: %+v", spans[0])
	}
	if got := []rune(first)[spans[0][1].From:spans[0][1].To]; string(got) != "[Open]" {
		t.Fatalf("the span of Open covers %q", string(got))
	}
}

func TestOnAShortLineTheButtonsGoUnderTheMetaLine(t *testing.T) {
	lines, spans := Head(plain, headInput(), 40)
	if strings.Contains(lines[0], "Run…") || !strings.Contains(pane.Strip(lines[1]), "[Run… (R)]  [Open]") {
		t.Fatalf("the buttons are not on their own line:\n%s", strings.Join(lines, "\n"))
	}
	if _, ok := spans[1]; !ok {
		t.Fatal("the line of buttons answers clicks")
	}
}

func TestTheTabsNameEverySectionAndEachTabAnswersAClick(t *testing.T) {
	for _, narrow := range []bool{false, true} {
		in := headInput()
		in.Narrow = narrow
		lines, spans := Head(plain, in, 80)
		row := -1
		for i, l := range lines {
			if strings.Contains(l, "Tasks") {
				row = i
			}
		}
		if row < 0 || len(spans[row]) != len(Sections) {
			t.Fatalf("narrow %v: no line of tabs:\n%s", narrow, strings.Join(lines, "\n"))
		}
		text := []rune(pane.Strip(lines[row]))
		for i, s := range Sections {
			sp := spans[row][i]
			if got := string(text[sp.From:sp.To]); got != s.Title(narrow) || sp.ID != s.ID() {
				t.Fatalf("narrow %v: tab %d covers %q with %q", narrow, i, got, sp.ID)
			}
		}
		if !strings.Contains(pane.Strip(lines[len(lines)-1]), "───") {
			t.Fatal("a rule ends the head")
		}
	}
	if Dependencies.Title(true) != "Deps" || Dependencies.Title(false) != "Dependencies" {
		t.Fatal("a narrow screen says Deps")
	}
}

func TestTheOpenTabHasTheLookOfTheActiveTab(t *testing.T) {
	was := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(was) })
	th := theme.Default()
	lines, _ := Head(th, headInput(), 80)
	tabs := lines[2]
	if !strings.Contains(tabs, th.ActiveTab("Runs")) || strings.Contains(tabs, th.ActiveTab("Tasks")) {
		t.Fatalf("only Runs is the active tab: %q", tabs)
	}
}
