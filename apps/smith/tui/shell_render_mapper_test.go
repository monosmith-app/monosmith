package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/splitter"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
)

func TestEveryScreenFillsExactlyTheTerminal(t *testing.T) {
	sizes := [][2]int{{140, 40}, {100, 30}, {99, 30}, {80, 24}, {79, 20}, {60, 15}, {40, 10}, {20, 5}, {10, 3}, {5, 2}, {1, 1}, {3, 1}}
	for _, inside := range []bool{true, false} {
		for _, sz := range sizes {
			l := newLab(t, inside, sz[0], sz[1])
			lines := strings.Split(l.sh().View(), "\n")
			if len(lines) != sz[1] {
				t.Fatalf("inside=%v %v: %d lines", inside, sz, len(lines))
			}
			for i, line := range lines {
				if w := pane.Width(line); w != sz[0] {
					t.Fatalf("inside=%v %v: line %d is %d cells wide: %q", inside, sz, i, w, line)
				}
			}
		}
	}
}

func TestAt140ColumnsTheHeaderTheListAndTheDetailShareTheScreen(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.hasNot("File ", "View ", "Help ", "Apps", "Libs") // no menu bar, no tree, no permanent side panel
	l.has("acme", "main",                               // the header
		"Projects", "Doctor", // the tabs
		"web", "orders", "kit", "sdk", "4 projects", // the list
		"root", "affected", "Build app", // the detail of the selected project
		"smith 1.2.3") // the status bar
	l.hasNot("Reading the workspace")
}

func TestAt100ColumnsTheDetailIsStillBesideTheList(t *testing.T) {
	newLab(t, true, 100, 30).has("root", "Build app", "4 projects")
}

func TestBelow100ColumnsOnlyTheListIsShown(t *testing.T) {
	l := newLab(t, true, 99, 30)
	l.has("4 projects", "web")
	l.hasNot("Build app", "language") // the detail opens with Enter
	newLab(t, true, 80, 26).hasNot("Build app")
}

func TestAtTheSupportedFloorEveryProjectIsStillListed(t *testing.T) {
	l := newLab(t, true, 72, 24)
	l.has("4 projects", "orders", "web", "kit", "sdk")
}

func TestOutsideAWorkspaceTheGuideIsTheOnlyTab(t *testing.T) {
	l := newLab(t, false, 100, 28)
	l.has("Start", "Create a workspace", "No workspace here")
	l.hasNot("Doctor", "Settings")
}

func TestTheInterfaceFollowsEveryResize(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.has("Build app")
	l.send(tea.WindowSizeMsg{Width: 90, Height: 20})
	l.hasNot("Build app")
	if got := len(strings.Split(l.sh().View(), "\n")); got != 20 {
		t.Fatalf("%d lines after the resize", got)
	}
	l.send(tea.WindowSizeMsg{Width: 72, Height: 20})
	l.has("4 projects")
	l.send(tea.WindowSizeMsg{Width: 140, Height: 34})
	l.has("Build app")
}

func TestAnOpenDetailClosesWhenTheScreenGrowsWideEnoughToShowItBesideTheList(t *testing.T) {
	l := newLab(t, true, 90, 30)
	l.keys("enter")
	if !l.sh().detailOpen {
		t.Fatal("Enter should open the detail")
	}
	l.send(tea.WindowSizeMsg{Width: 130, Height: 30})
	if l.sh().detailOpen {
		t.Fatal("the detail is beside the list now")
	}
	l.has("Build app", "4 projects")
}

func TestTheFirstProjectIsSelectedAndShownInTheDetail(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.has("orders", "root", "go", "Run on orders", "Build app", "turns the source into what gets shipped", "Test app", "How was this resolved?")
}

func TestALongPathInTheDetailIsShortenedInTheMiddle(t *testing.T) {
	l := newLab(t, true, 100, 30)
	sh := l.sh()
	sh.list = projectlist.New([]projectlist.Project{{Name: "far", Kind: "app", Language: "go", Root: "apps/some/deeply/nested/place/far-away-project"}})
	l.m = sh
	l.has("apps/", "project", "…")
	l.hasNot("far-away-project")
}

func TestNoColourMeansSymbolsAndWordsStillCarryTheMeaning(t *testing.T) {
	// The lab uses the plain theme: no colour codes except the underlined mnemonics.
	raw := newLab(t, true, 120, 30).sh().View()
	if strings.Contains(raw, "\x1b[38;") || strings.Contains(raw, "\x1b[48;") {
		t.Fatalf("colour in the plain theme: %q", raw[:200])
	}
}

// Rendering must be total: no state or size may panic, including the last
// frame Bubble Tea draws after quitting.
func TestRenderingNeverPanicsAfterAnyTransition(t *testing.T) {
	events := []any{
		"down", "down", "right", "enter", "esc", "left", "up", "home", "end", "esc", "j", "k", "/", "esc",
		"ctrl+p", "d", "o", "enter", "esc", "esc", "f1", "down", "pgdown", "esc", ":", "esc", "?", "esc",
		"tab", "tab", "tab", "shift+tab", "right", "left", "space", "enter", "1", "2", "3", "4", "8", "9",
		"3", "x", "esc", "n", "a", "esc", "b", "f", "r", "pgup", "pgdown", "ctrl+c", "enter", "i", "y", "4", "p", "x", "esc", "m", "m",
		"5", "e", "6", "u", "v", "space", "7", "c", "esc",
		click(3, 0), click(10, 3), click(40, 2), click(5, 8), rightClick(5, 4), motion(9, 9), wheel(10, 10, 1), wheel(40, 10, -1), click(0, 0),
		"q",
	}
	for _, sz := range [][2]int{{140, 40}, {100, 30}, {80, 24}, {60, 20}, {20, 5}, {1, 1}} {
		for _, inside := range []bool{true, false} {
			l := newLab(t, inside, sz[0], sz[1])
			l.next = &scripted{lines: []string{"kit:build ▶ go build", "kit:build │ hello", "kit:build × exited with code 2"}, exit: runner.Exit{Code: 1}}
			for i, ev := range events {
				switch e := ev.(type) {
				case string:
					l.keys(e)
				default:
					l.send(e.(tea.Msg))
				}
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Fatalf("size %v inside=%v: panic after event %d (%v): %v", sz, inside, i, ev, r)
						}
					}()
					lines := l.sh().screen()
					if len(lines) != sz[1] {
						t.Fatalf("size %v after %v: %d lines", sz, ev, len(lines))
					}
				}()
			}
			_ = l.sh().View() // the frame after quitting
		}
	}
}

func TestClickingEveryCellInEveryStateNeverPanics(t *testing.T) {
	prepare := map[string]func(*lab){
		"idle":           func(*lab) {},
		"palette":        func(l *lab) { l.keys("ctrl+p") },
		"help":           func(l *lab) { l.keys("f1") },
		"confirm":        func(l *lab) { l.keys("5", "x", "enter") },
		"release choice": func(l *lab) { l.keys("5", "x") },
		"run": func(l *lab) {
			l.next = &scripted{lines: []string{"kit:build ▶ x", "kit:build ✓ 1s", "x:y │ z"}}
			l.keys("b")
		},
		"home":   func(l *lab) { l.keys("1") },
		"doctor": func(l *lab) { l.keys("3") },
		"deps":   func(l *lab) { l.keys("4") },
		"guide":  func(l *lab) { l.keys("a") },
	}
	for name, prep := range prepare {
		for _, btn := range []func(x, y int) any{func(x, y int) any { return click(x, y) }, func(x, y int) any { return rightClick(x, y) }, func(x, y int) any { return wheel(x, y, 1) }} {
			l := newLab(t, true, 80, 24)
			prep(l)
			for y := 0; y < 24; y++ {
				for x := 0; x < 80; x += 3 {
					func() {
						defer func() {
							if r := recover(); r != nil {
								t.Fatalf("%s: panic at (%d,%d): %v\n%s", name, x, y, r, l.view())
							}
						}()
						l.send(btn(x, y).(tea.Msg))
						_ = l.sh().screen()
					}()
				}
			}
		}
	}
}

// ---- the main screen's frame and the dimming behind a modal ----------------

// frameTabs set up each tab with a frame: Projects, Run (a run is open), Doctor,
// Settings and Home.
var frameTabs = []struct {
	name string
	set  func(t *testing.T, w, h int) *lab
}{
	{"projects", func(t *testing.T, w, h int) *lab { return newLab(t, true, w, h).openTab(commands.TabProjects) }},
	{"run", func(t *testing.T, w, h int) *lab {
		l := newLab(t, true, w, h)
		l.next = failingRun()
		l.keys("b")
		// The toast of the finished run is a box with corners, which is not the frame.
		sh := l.sh()
		sh.toast = toast.Toast{}
		l.m = sh
		return l
	}},
	{"doctor", func(t *testing.T, w, h int) *lab { return newLab(t, true, w, h).openTab(commands.TabDoctor) }},
	{"settings", func(t *testing.T, w, h int) *lab {
		l, _ := newSettingsLab(t, w, h)
		return l.keys("6")
	}},
	{"home", func(t *testing.T, w, h int) *lab { return newHomeLab(t, w, h) }},
}

// frameTab is the lab of the tab called name at w by h cells.
func frameTab(t *testing.T, name string, w, h int) *lab {
	t.Helper()
	for _, f := range frameTabs {
		if f.name == name {
			return f.set(t, w, h)
		}
	}
	t.Fatalf("no tab %q", name)
	return nil
}

func TestTheMainScreenIsARuleUnderTheTabsWithNoSidesAndNoBottom(t *testing.T) {
	for _, w := range []int{80, 120} {
		for _, tc := range frameTabs {
			l := tc.set(t, w, 40)
			lay := l.sh().lay()
			lines := l.lines()
			for y := lay.Main.Y; y < lay.Footer.Y; y++ {
				row := []rune(lines[y])
				if len(row) != w {
					t.Fatalf("%s at %d: row %d is %d cells", tc.name, w, y, len(row))
				}
				if strings.ContainsAny(lines[y], "╭╮╰╯┴") {
					t.Errorf("%s at %d: row %d has a corner or a bottom tee: %q", tc.name, w, y, lines[y])
				}
				if first, last := row[0], row[w-1]; (first != ' ' && first != '─') || (last != ' ' && last != '─') {
					t.Errorf("%s at %d: row %d starts or ends with a border glyph: %q", tc.name, w, y, lines[y])
				}
			}
			// The rule is one line across the width; a tee marks each vertical division.
			tees := map[int]bool{}
			for _, j := range splitter.Junctions(l.sh().spots(), lay.Main.Y) {
				if j.Y == lay.Main.Y {
					tees[j.X] = true
				}
			}
			if lay.HasDetail() && !tees[lay.Detail.X-1] {
				t.Errorf("%s at %d: no tee where the detail's divider meets the rule", tc.name, w)
			}
			for x, r := range []rune(lines[lay.Main.Y]) {
				if want := map[bool]rune{true: '┬', false: '─'}[tees[x]]; r != want {
					t.Errorf("%s at %d: the rule has %q at column %d, want %q: %q", tc.name, w, r, x, want, lines[lay.Main.Y])
					break
				}
			}
		}
	}
}

func TestEveryDividerIsHeldAndDrawnHeavyOnItsNewCellsAt80And120Columns(t *testing.T) {
	for _, w := range []int{80, 120} {
		for _, tc := range frameTabs {
			for i := 0; ; i++ {
				l := tc.set(t, w, 40)
				spots := l.sh().spots()
				if i >= len(spots) {
					break
				}
				sp := spots[i]
				x, y, rest, heavy, nx, ny := sp.Hit.X, sp.Hit.Y+1, '│', '┃', sp.Hit.X-2, sp.Hit.Y+1
				if sp.Axis == splitter.Row {
					// Near the right end: the rule of the run log carries its title in the middle.
					x, y, rest, heavy, nx, ny = sp.Hit.X+sp.Hit.W-3, sp.Hit.Y, '─', '━', sp.Hit.X+sp.Hit.W-3, sp.Hit.Y-1
				}
				cell := func() rune { return []rune(l.lines()[y])[x] }
				if cell() != rest {
					t.Fatalf("%s at %d: the %s divider is drawn %q at (%d,%d), want %q", tc.name, w, sp.ID, cell(), x, y, rest)
				}
				l.send(click(x, y))
				if l.sh().split.Dragging != sp.ID || cell() != heavy {
					t.Fatalf("%s at %d: pressing the %s divider at (%d,%d) holds %q and draws %q", tc.name, w, sp.ID, x, y, l.sh().split.Dragging, cell())
				}
				l.send(drag(nx, ny)).send(letGo(nx, ny))
				if n, ok := l.sh().split.Size(sp.ID); !ok || n != sp.SizeAt(nx, ny) {
					t.Errorf("%s at %d: dragging the %s divider to (%d,%d) sized it %d, want %d", tc.name, w, sp.ID, nx, ny, n, sp.SizeAt(nx, ny))
				}
			}
		}
	}
}

// marked is true for a row the list shows as selected: in reverse video or with the pointer.
func marked(row string) bool { return strings.Contains(row, "7m") || strings.Contains(row, "›") }

func TestAClickInAListSelectsTheRowUnderItAt80And120Columns(t *testing.T) {
	for _, w := range []int{80, 120} {
		for _, tc := range []struct{ tab, text string }{{"projects", ""}, {"run", ""}, {"doctor", "Docker registries"}, {"settings", "Recipes"}} {
			l := frameTab(t, tc.tab, w, 40)
			text := tc.text
			switch tc.tab {
			case "projects":
				text = l.sh().list.Projects()[1].Name
			case "run":
				sh := l.sh()
				text = sh.runIDs()[2]
			}
			x, y := l.find(text)
			if marked(l.rawRow(y)) {
				t.Fatalf("%s at %d: %q is selected before the click", tc.tab, w, text)
			}
			l.send(click(x+1, y)).send(letGo(x+1, y))
			if !marked(l.rawRow(y)) {
				t.Errorf("%s at %d: a click on %q at (%d,%d) did not select its row:\n%s", tc.tab, w, text, x+1, y, l.view())
			}
		}
	}
}

func TestADragSelectsTheTextUnderItAt80And120Columns(t *testing.T) {
	for _, w := range []int{80, 120} {
		for _, tab := range []string{"projects", "doctor"} {
			l := frameTab(t, tab, w, 40)
			out := copying(l)
			text := "Docker registries"
			if tab == "projects" {
				text = l.sh().list.Projects()[1].Name
			}
			x, y := l.find(text)
			l.send(click(x, y)).send(drag(x+2, y)).send(letGo(x+2, y))
			if got := l.copied(t, out); got != text[:3] {
				t.Errorf("%s at %d: copied %q, want %q", tab, w, got, text[:3])
			}
		}
	}
}

func TestTheWheelScrollsTheListFromItsLastRowAt80And120Columns(t *testing.T) {
	for _, w := range []int{80, 120} {
		l := newLab(t, true, w, 30)
		sh := l.sh()
		var rows []depsview.Row
		for i := 0; i < 60; i++ {
			rows = append(rows, depsview.Row{Name: fmt.Sprintf("pkg%02d", i), Current: "1.0.0", Latest: "1.0.1", Level: "patch", Projects: []string{"web"}})
		}
		sh.sources.Deps = func() depsview.Result { return depsview.Result{Rows: rows} }
		l.m = sh
		l.keys("4", "r")
		last := l.sh().mainInner().Y + l.sh().mainInner().H - 1
		if last != l.sh().lay().Footer.Y-1 {
			t.Fatalf("at %d the content ends on row %d, the footer is on %d", w, last, l.sh().lay().Footer.Y)
		}
		before := l.view()
		l.send(wheel(4, last, 1)).send(wheel(4, last, 1))
		if l.view() == before {
			t.Errorf("at %d the wheel on the last content row did not scroll the list", w)
		}
	}
}

func TestBehindAModalTheWholeScreenIsDimmedAndWithoutOneItIsNot(t *testing.T) {
	l := colourLab(t, 120, 40, commands.TabProjects)
	lay, th := l.sh().lay(), l.sh().th
	rows := []int{0, 1, lay.Main.Y, lay.Footer.Y, lay.StatusBar.Y}
	for _, y := range rows {
		if raw := l.rawRow(y); raw == th.Faint(pane.Strip(raw)) {
			t.Errorf("row %d is dimmed without a modal: %q", y, raw)
		}
	}
	if raw := l.rawRow(1); !strings.Contains(raw, ";48;") {
		t.Fatalf("the tab buttons have a background colour: %q", raw)
	}
	l.keys("f1")
	for _, y := range rows {
		raw := l.rawRow(y)
		if want := th.Faint(pane.Strip(raw)); raw != want || !strings.Contains(raw, "\x1b[") || strings.Contains(raw, "48;") {
			t.Errorf("row %d behind the modal is not the faint style alone: %q", y, raw)
		}
	}
	l.keys("esc")
	if raw := l.rawRow(1); !strings.Contains(raw, "48;") {
		t.Fatalf("the screen is bright again once the modal closes: %q", raw)
	}
}

func TestInThePlainThemeTheDimmingIsTheFaintAttributeAndOnlyBehindAModal(t *testing.T) {
	l := newLab(t, true, 120, 40)
	if strings.Contains(l.rawRow(1), "\x1b[2m") {
		t.Fatal("no faint attribute without a modal")
	}
	l.keys("f1")
	if !strings.HasPrefix(l.rawRow(1), "\x1b[2m") {
		t.Fatalf("the plain theme dims with the faint attribute: %q", l.rawRow(1))
	}
}

func TestTheJunctionTakesTheColourOfItsDividerWhileItIsSelectedOrHeld(t *testing.T) {
	l := doctorLab(t, 140, 30)
	lay := l.sh().lay()
	x, rule := lay.Detail.X-1, lay.Main.Y
	th := l.sh().th
	rest, accent, active := th.Border("┬", false), th.Accent("┬"), th.ButtonActive("┬")
	cell := func() string { return cellAt(l.rawRow(rule), x) }
	if cell() != rest {
		t.Fatalf("at rest the tee is in the rule's colour: %q, want %q", cell(), rest)
	}
	l.send(tea.KeyMsg{Type: tea.KeyShiftRight})
	if cell() != accent {
		t.Fatalf("selected: the tee is in the accent: %q, want %q", cell(), accent)
	}
	l.keys("esc")
	if cell() != rest {
		t.Fatalf("let go of by the keys: the tee returns to the rule's colour: %q", cell())
	}
	l.send(click(x, rule+4))
	if cell() != active {
		t.Fatalf("held: the tee is in the held colour: %q, want %q", cell(), active)
	}
	l.send(letGo(3, rule+9))
	if cell() != rest {
		t.Fatalf("released: the tee returns to the rule's colour: %q, want %q", cell(), rest)
	}
}

// cellAt is the cell at column x of a styled row, with its escape codes.
func cellAt(row string, x int) string { return ansi.Cut(row, x, x+1) }
