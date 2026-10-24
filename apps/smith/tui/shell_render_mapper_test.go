package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
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
		"web", "orders", "kit", "sdk", "Projects (4)", // the list
		"kind", "language", "Run build", // the detail of the selected project
		"smith 1.2.3") // the status bar
	l.hasNot("Reading the workspace")
}

func TestAt100ColumnsTheDetailIsStillBesideTheList(t *testing.T) {
	newLab(t, true, 100, 30).has("kind", "Run build", "On: orders")
}

func TestBelow100ColumnsOnlyTheListIsShown(t *testing.T) {
	l := newLab(t, true, 99, 30)
	l.has("On: orders", "web")
	l.hasNot("Run build", "language") // the detail opens with Enter
	newLab(t, true, 80, 26).hasNot("Run build")
}

func TestAtTheSupportedFloorEveryProjectIsStillListed(t *testing.T) {
	l := newLab(t, true, 72, 24)
	l.has("On: orders", "orders", "web", "kit", "sdk")
}

func TestOutsideAWorkspaceTheGuideIsTheOnlyTab(t *testing.T) {
	l := newLab(t, false, 100, 28)
	l.has("Guide", "Create a workspace", "No workspace here")
	l.hasNot("Doctor", "Settings")
}

func TestTheInterfaceFollowsEveryResize(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.has("Run build")
	l.send(tea.WindowSizeMsg{Width: 90, Height: 20})
	l.hasNot("Run build")
	if got := len(strings.Split(l.sh().View(), "\n")); got != 20 {
		t.Fatalf("%d lines after the resize", got)
	}
	l.send(tea.WindowSizeMsg{Width: 72, Height: 20})
	l.has("On: orders")
	l.send(tea.WindowSizeMsg{Width: 140, Height: 34})
	l.has("Run build")
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
	l.has("Run build", "On: orders")
}

func TestTheFirstProjectIsSelectedAndShownInTheDetail(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.has("orders", "kind", "language", "go", "Actions", "Run build", "Run test", "Explain")
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
			l.next = &scripted{lines: []string{"kit:build ▶ go build", "kit:build │ hello", "kit:build ✖ exited with code 2"}, exit: runner.Exit{Code: 1}}
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
		"idle":    func(*lab) {},
		"palette": func(l *lab) { l.keys("ctrl+p") },
		"help":    func(l *lab) { l.keys("f1") },
		"confirm": func(l *lab) { l.keys("6", "x") },
		"run": func(l *lab) {
			l.next = &scripted{lines: []string{"kit:build ▶ x", "kit:build ✔ 1s", "x:y │ z"}}
			l.keys("b")
		},
		"home":   func(l *lab) { l.keys("1") },
		"doctor": func(l *lab) { l.keys("4") },
		"deps":   func(l *lab) { l.keys("5") },
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
