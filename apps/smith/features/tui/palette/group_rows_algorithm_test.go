package palette

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
)

func TestTheGroupsHaveTheirCountsAndEveryEntryShowsItsCommandLine(t *testing.T) {
	entries := []Entry{
		{ID: "run.build", Section: "Build", Title: "Build web", Hint: "smith run build -p web", Shortcut: "b"},
		{ID: "run.test", Section: "Build", Title: "Test web", Hint: "smith run test -p web", Shortcut: "t"},
		{ID: "file.add", Section: "Projects", Title: "Add project…", Hint: "smith add: an app, lib or package", Shortcut: "a"},
	}
	lines := New(entries).Lines(theme.Default(), 70, 14)
	plain := make([]string, len(lines))
	for i, l := range lines {
		plain[i] = strings.TrimRight(pane.Strip(l), " ")
	}
	text := strings.Join(plain, "\n")
	for _, want := range []string{" Build  2", " Projects  1", "    smith run build -p web", "    smith run test -p web", "    smith add"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
	if strings.Contains(text, "an app, lib or package") {
		t.Fatalf("the second line is the command line, not its description:\n%s", text)
	}
	if last := plain[len(plain)-1]; !strings.Contains(last, "↑/↓ move") || !strings.Contains(last, "esc close") {
		t.Fatalf("the last line has the key hints: %q", last)
	}
}

func TestASearchPutsTheBestMatchFirstThenTheOthersBySectionAndTheUnusableLast(t *testing.T) {
	entries := []Entry{
		{ID: "a", Section: "Health", Title: "Doctor web", Hint: "smith doctor"},
		{ID: "b", Section: "Build", Title: "Build web", Hint: "smith run build -p web"},
		{ID: "c", Section: "Build", Title: "Stop web", Hint: "cancel", Unavailable: "nothing runs"},
		{ID: "d", Section: "Health", Title: "Lint web", Hint: "smith run lint -p web"},
	}
	p := typed(New(entries), "web")
	var order []string
	for _, e := range p.matches() {
		order = append(order, e.ID)
	}
	if got := strings.Join(order, ","); got != "a,b,d,c" {
		t.Fatalf("order %s", got)
	}
	text := pane.Strip(strings.Join(p.Lines(theme.Default(), 60, 20), "\n"))
	for _, want := range []string{"Best match  1", "Build  1", "Health  1", "Not usable now  1", "nothing runs"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
}

func TestTheNarrowPaletteFitsAndAClickOnEitherLineOfAnEntryChoosesIt(t *testing.T) {
	// The palette of an 80x30 screen: 72 cells, 22 lines inside its frame.
	p := New(ui)
	lines := p.Lines(theme.Default(), 70, 22)
	if len(lines) != 22 {
		t.Fatalf("%d lines", len(lines))
	}
	for _, l := range lines {
		if pane.Width(l) != 70 {
			t.Fatalf("%q is %d wide", pane.Strip(l), pane.Width(l))
		}
	}
	// Line 2 is the Build heading, 3 and 4 are Build web.
	if p.At(2, 22) != -1 || p.At(3, 22) != 0 || p.At(4, 22) != 0 || p.At(0, 22) != -1 || p.At(21, 22) != -1 {
		t.Fatalf("at: %d %d %d", p.At(2, 22), p.At(3, 22), p.At(4, 22))
	}
}

func TestTheWindowKeepsBothLinesOfTheSelectedEntry(t *testing.T) {
	var many []Entry
	for i := 0; i < 20; i++ {
		many = append(many, Entry{ID: string(rune('a' + i)), Section: "S", Title: "Entry " + string(rune('a'+i)), Hint: "smith x"})
	}
	p := New(many)
	for i := 0; i < 10; i++ {
		p, _ = p.Key(tea.KeyMsg{Type: tea.KeyDown})
	}
	lines := p.Lines(theme.Default(), 40, 12)
	found := false
	for i, l := range lines {
		if strings.Contains(pane.Strip(l), selectedMark+"Entry k") {
			found = true
			if !strings.Contains(pane.Strip(lines[i+1]), "smith x") {
				t.Fatalf("its command line is cut off:\n%s", pane.Strip(strings.Join(lines, "\n")))
			}
		}
	}
	if !found {
		t.Fatalf("the selected entry is out of view:\n%s", pane.Strip(strings.Join(lines, "\n")))
	}
}
