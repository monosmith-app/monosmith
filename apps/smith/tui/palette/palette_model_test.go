package palette

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

var cli = []clicontract.CommandDescription{
	{Name: "affected", Description: "List projects a change touches, before a release"},
	{Name: "doctor", Description: "Check the workspace"},
	{Name: "release", Description: "Version, tag and publish", Options: []clicontract.OptionDescription{{Name: "dry-run"}}},
}

var ui = []Entry{
	{ID: "run.build", Title: "Build", Hint: "smith build", Shortcut: "Ctrl+B"},
	{ID: "run.stop", Title: "Stop", Hint: "cancel the run", Shortcut: "Ctrl+C", Unavailable: "nothing is running"},
}

func typed(p Palette, text string) Palette {
	for _, r := range text {
		p, _ = p.Key(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return p
}

func enter() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyEnter} }

func TestTheListHasEveryShellCommandAndEverySmithCommand(t *testing.T) {
	text := strings.Join(New(ui, cli).Lines(theme.Default(), 70, 12), "\n")
	for _, want := range []string{"Build", "Ctrl+B", "affected", "doctor", "release"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
}

func TestTypingFindsASmithCommandAndEnterAsksForItsArguments(t *testing.T) {
	p := typed(New(ui, cli), "rel")
	if text := strings.Join(p.Lines(theme.Default(), 70, 12), "\n"); !strings.Contains(text, "release") {
		t.Fatalf("release should match:\n%s", text)
	}
	p, res := p.Key(enter())
	if res.Closed || p.Chosen() == nil || p.Chosen().Name != "release" {
		t.Fatalf("%+v", res)
	}
	p = typed(p, "--dry-run")
	_, res = p.Key(enter())
	if strings.Join(res.Argv, " ") != "release --dry-run" || !res.Closed {
		t.Fatalf("%+v", res)
	}
}

func TestPickingAShellCommandReportsItsID(t *testing.T) {
	_, res := typed(New(ui, cli), "build").Key(enter())
	if res.CommandID != "run.build" || !res.Closed {
		t.Fatalf("%+v", res)
	}
}

func TestAnUnavailableShellCommandSaysWhyAndStaysOpen(t *testing.T) {
	_, res := typed(New(ui, cli), "stop").Key(enter())
	if res.CommandID != "" || res.Closed || !strings.Contains(res.Refused, "nothing is running") {
		t.Fatalf("%+v", res)
	}
}

func TestEscBacksOutOfArgumentsThenClosesThePalette(t *testing.T) {
	p, _ := typed(New(ui, cli), "doctor").Key(enter())
	p, res := p.Key(tea.KeyMsg{Type: tea.KeyEsc})
	if res.Closed || p.Chosen() != nil {
		t.Fatal("first esc goes back")
	}
	if _, res = p.Key(tea.KeyMsg{Type: tea.KeyEsc}); !res.Closed {
		t.Fatal("second esc closes")
	}
}

func TestClickingARowChoosesIt(t *testing.T) {
	p := New(ui, cli)
	i := p.RowAt(0, 9)
	if i != 0 {
		t.Fatalf("row 0 is match %d", i)
	}
	_, res := p.Choose(i)
	if res.CommandID != "run.build" {
		t.Fatalf("%+v", res)
	}
	if p.RowAt(50, 9) != -1 {
		t.Fatal("below the list")
	}
}

func TestTheWheelAndTheArrowsMoveTheCursor(t *testing.T) {
	p := New(ui, cli).Scroll(2)
	if p.cursor != 2 {
		t.Fatalf("%d", p.cursor)
	}
	p, _ = p.Key(tea.KeyMsg{Type: tea.KeyUp})
	if p.cursor != 1 {
		t.Fatalf("%d", p.cursor)
	}
	if New(ui, cli).Scroll(99).cursor != 3 || New(ui, cli).Scroll(-9).cursor != 0 {
		t.Fatal("clamped")
	}
}

func TestRenderingNeverPanicsAtAnySize(t *testing.T) {
	for _, s := range [][2]int{{0, 0}, {1, 1}, {5, 2}, {20, 5}, {80, 24}} {
		p := New(ui, cli)
		_ = p.Lines(theme.Default(), s[0], s[1])
		p = typed(p, "zzzz")
		_ = p.Lines(theme.Default(), s[0], s[1])
		p, _ = New(ui, cli).Choose(2)
		_ = p.Lines(theme.Default(), s[0], s[1])
	}
}
