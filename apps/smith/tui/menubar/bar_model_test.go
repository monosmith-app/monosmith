package menubar

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func sample() Bar {
	return New([]Menu{
		{Name: "File", Entries: []Entry{
			{ID: "palette", Name: "Command palette", Shortcut: "Ctrl+P"},
			{Separator: true},
			{ID: "quit", Name: "Quit", Shortcut: "Ctrl+Q"},
		}},
		{Name: "Run", Entries: []Entry{
			{ID: "run", Name: "Run selection", Shortcut: "Ctrl+R"},
			{ID: "stop", Name: "Stop", Shortcut: "Ctrl+C", Disabled: "nothing is running"},
		}},
		{Name: "Release", Mnemonic: 'e', Entries: []Entry{{ID: "preview", Name: "Preview", Shortcut: "F2"}}},
	}, 80)
}

func TestF10OpensTheFirstMenuAndTheArrowsWalkTheEntries(t *testing.T) {
	b, r := sample().Key("f10")
	if !r.Handled || b.Open() != 0 || b.Cursor() != 0 {
		t.Fatalf("open %d cursor %d", b.Open(), b.Cursor())
	}
	b, _ = b.Key("down")
	if b.Cursor() != 2 {
		t.Fatalf("down should skip the separator: %d", b.Cursor())
	}
	b, _ = b.Key("down")
	if b.Cursor() != 0 {
		t.Fatalf("down wraps: %d", b.Cursor())
	}
	b, r = b.Key("right")
	if b.Open() != 1 || !r.Handled {
		t.Fatalf("right opens the next menu: %d", b.Open())
	}
	b, _ = b.Key("esc")
	if b.Open() != -1 {
		t.Fatal("esc closes")
	}
}

func TestAltAndTheMnemonicOpenAMenu(t *testing.T) {
	b, _ := sample().Key("alt+r")
	if b.Open() != 1 {
		t.Fatalf("alt+r opens Run, got %d", b.Open())
	}
	b, _ = sample().Key("alt+e")
	if b.Open() != 2 {
		t.Fatalf("Release has the mnemonic e, got %d", b.Open())
	}
	if _, r := sample().Key("alt+z"); r.Handled {
		t.Fatal("an unknown letter is not ours")
	}
}

func TestEnterChoosesAndReportsTheEntryID(t *testing.T) {
	b, _ := sample().Key("alt+f")
	b, _ = b.Key("down")
	b, r := b.Key("enter")
	if r.Chosen != "quit" || b.Open() != -1 {
		t.Fatalf("chosen %q open %d", r.Chosen, b.Open())
	}
}

func TestADisabledEntryIsNotChosenAndSaysWhy(t *testing.T) {
	b, _ := sample().Key("alt+r")
	b, _ = b.Key("down")
	if !strings.Contains(b.Tooltip(), "nothing is running") {
		t.Fatalf("tooltip %q", b.Tooltip())
	}
	b, r := b.Key("enter")
	if r.Chosen != "" || !strings.Contains(r.Refused, "nothing is running") || b.Open() != 1 {
		t.Fatalf("result %+v open %d", r, b.Open())
	}
}

func TestClickingAHeadingOpensItAndClickingAgainCloses(t *testing.T) {
	b := sample()
	x := b.headingX()[1] + 1
	b, r := b.Click(x, 0)
	if b.Open() != 1 || !r.Handled {
		t.Fatalf("open %d", b.Open())
	}
	b, _ = b.Click(x, 0)
	if b.Open() != -1 {
		t.Fatal("a second click closes")
	}
}

func TestClickingAnEntryChoosesIt(t *testing.T) {
	b := sample().OpenMenu(0)
	d := b.box()
	// Entry 0 is on row Y+1, inside the left border.
	b, r := b.Click(d.X+2, d.Y+1)
	if r.Chosen != "palette" || b.Open() != -1 {
		t.Fatalf("chosen %q", r.Chosen)
	}
}

func TestClickingOutsideClosesWithoutChoosing(t *testing.T) {
	b := sample().OpenMenu(0)
	b, r := b.Click(60, 15)
	if r.Chosen != "" || b.Open() != -1 {
		t.Fatalf("result %+v open %d", r, b.Open())
	}
}

func TestHoveringAnotherHeadingSwitchesMenusAndHoveringAnEntryHighlightsIt(t *testing.T) {
	b := sample().OpenMenu(0)
	b = b.Hover(sample().headingX()[1]+1, 0)
	if b.Open() != 1 {
		t.Fatalf("hover switches menu: %d", b.Open())
	}
	d := b.box()
	b = b.Hover(d.X+2, d.Y+2)
	if b.Cursor() != 1 {
		t.Fatalf("cursor %d", b.Cursor())
	}
}

func TestEveryEntryShowsItsShortcut(t *testing.T) {
	b := sample().OpenMenu(0)
	lines, _, _, ok := b.Box(theme.Default())
	if !ok {
		t.Fatal("no box")
	}
	text := strings.Join(lines, "\n")
	for _, want := range []string{"Command palette", "Ctrl+P", "Quit", "Ctrl+Q"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, text)
		}
	}
}

func TestThePopupChoosesWithAClickOrAKey(t *testing.T) {
	entries := []Entry{{ID: "run", Name: "Run"}, {ID: "why", Name: "Why a miss", Disabled: "no cache entry"}}
	p := OpenPopup(entries, 10, 5, 80, 24)
	if !p.Shown() {
		t.Fatal("not shown")
	}
	p, _ = p.Key("down")
	if !strings.Contains(p.Tooltip(), "no cache entry") {
		t.Fatalf("tooltip %q", p.Tooltip())
	}
	if _, r := p.Key("enter"); r.Chosen != "" || r.Refused == "" {
		t.Fatalf("%+v", r)
	}
	p = OpenPopup(entries, 10, 5, 80, 24)
	_, r := p.Click(p.box.X+2, p.box.Y+1)
	if r.Chosen != "run" {
		t.Fatalf("chosen %q", r.Chosen)
	}
	if p, _ = p.Click(70, 20); p.Shown() {
		t.Fatal("a click outside closes it")
	}
}

func TestThePopupStaysOnScreen(t *testing.T) {
	p := OpenPopup([]Entry{{ID: "a", Name: "A long entry name", Shortcut: "Ctrl+X"}}, 78, 23, 80, 24)
	if p.box.X+p.box.innerWidth()+2 > 80 || p.box.Y+p.box.height() > 24 {
		t.Fatalf("box at %d,%d is off a 80x24 screen", p.box.X, p.box.Y)
	}
}

func TestTheAppNameAndAPipeComeBeforeTheMenus(t *testing.T) {
	b := sample().SetBrand("MonoSmith")
	row := pane.Strip(b.Row(theme.Default(), "", 60))
	if !strings.HasPrefix(row, " MonoSmith │ File") {
		t.Fatalf("row %q", row)
	}
	x := strings.Index(row, "File")
	if b.HeadingAt(x) != 0 || b.HeadingAt(x-4) != -1 {
		t.Fatalf("headings must be found after the name: %d %d", b.HeadingAt(x), b.HeadingAt(x-4))
	}
}
