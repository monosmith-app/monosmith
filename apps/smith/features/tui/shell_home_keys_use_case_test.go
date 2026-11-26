package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/dashboard"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runview"
)

const (
	moreBelow = "↓ more (PgDn, Ctrl+D)"
	moreAbove = "↑ more (PgUp, Ctrl+U)"
)

// homeItems are the items of the list of Home.
func homeItems(l *lab) []dashboard.Rows {
	sh := l.sh()
	return sh.homePage(sh.mainInner().W, dashboard.Away).Items
}

// showOnHome pages Home down until text is on screen.
func showOnHome(l *lab, text string) *lab {
	for i := 0; i < 40 && !strings.Contains(l.view(), text); i++ {
		l.keys("pgdown")
	}
	return l
}

// finishedRun puts a run that ended in the history of the lab.
func finishedRun(l *lab, argv ...string) {
	sh := l.sh()
	r := runview.NewRun(argv, nil, l.clock)
	r.End(0, false, l.clock)
	sh.runs = append(sh.runs, r)
	l.m = sh
}

func TestAShortHomeScrollsAndSaysWhatIsCut(t *testing.T) {
	l := newHomeLab(t, 120, 14)
	l.has(moreBelow, "4 projects", "Changed")
	l.hasNot("Compose a stack", moreAbove)
	for i := 0; i < 8 && !strings.Contains(l.view(), "Compose a stack"); i++ {
		l.keys("pgdown")
		l.has(moreAbove)
	}
	l.has("Compose a stack", "Add a project")
	l.hasNot(moreBelow)
	for range 8 {
		l.keys("pgup")
	}
	l.has("4 projects", moreBelow)
	l.hasNot(moreAbove)
}

func TestCtrlDAndCtrlUAndTheWheelScrollHome(t *testing.T) {
	l := newHomeLab(t, 120, 14)
	l.keys("ctrl+d")
	l.has(moreAbove)
	l.hasNot("4 projects (2 Go")
	l.keys("ctrl+u")
	l.has("4 projects (2 Go")
	l.hasNot(moreAbove)
	x, y := l.find("4 projects")
	l.send(wheel(x, y, 1))
	if l.sh().homeTop == 0 {
		t.Fatal("the wheel did not scroll Home")
	}
	l.send(wheel(x, y, -1))
	if l.sh().homeTop != 0 {
		t.Fatalf("the wheel up left Home at line %d", l.sh().homeTop)
	}
	// The wheel never moves the keyboard.
	if l.sh().homeSel != 0 || l.sh().homeBtn != 0 {
		t.Fatalf("the wheel moved the keyboard to %d/%d", l.sh().homeSel, l.sh().homeBtn)
	}
}

func TestTheArrowsMoveThroughTheItemsAndTheirButtons(t *testing.T) {
	l := newHomeLab(t, 160, 44)
	l.has("[Fix (Enter)]", "[Open Health (3)]", "Needs you now  1/4")
	l.keys("down")
	l.has("[Fix (F)]", "[Test what changed (Enter)]", "Needs you now  2/4")
	l.keys("up", "right")
	l.has("[Fix (F)]", "[Open Health (Enter)]")
	if sh := l.sh(); sh.homeSel != 0 || sh.homeBtn != 1 {
		t.Fatalf("Right moved to %d/%d", sh.homeSel, sh.homeBtn)
	}
	l.keys("right")
	if l.sh().homeBtn != 1 {
		t.Fatalf("past the last button the keyboard stays: %d", l.sh().homeBtn)
	}
	l.keys("left", "left", "up")
	if sh := l.sh(); sh.homeSel != 0 || sh.homeBtn != 0 {
		t.Fatalf("Left and Up stop at the first button of the first item: %d/%d", sh.homeSel, sh.homeBtn)
	}
	l.keys("right", "down")
	if l.sh().homeBtn != 0 {
		t.Fatal("a new item starts on its first button")
	}
	l.keys("end")
	last := len(homeItems(l)) - 1
	if l.sh().homeSel != last {
		t.Fatalf("End went to %d, want %d", l.sh().homeSel, last)
	}
	l.keys("down")
	if l.sh().homeSel != last {
		t.Fatalf("past the last item the keyboard stays: %d", l.sh().homeSel)
	}
	l.has("[Open Deps (Enter)]", "Needs you now  4/4")
	l.keys("j", "k", "home")
	if l.sh().homeSel != 0 {
		t.Fatalf("Home went to %d", l.sh().homeSel)
	}
}

func TestEnterPressesTheButtonTheKeyboardIsOnAndTheFooterNamesIt(t *testing.T) {
	l := newHomeLab(t, 160, 44)
	footer := func() string { lines := l.lines(); return lines[len(lines)-2] }
	if !strings.Contains(footer(), "Enter fix") {
		t.Fatalf("the footer names the action: %q", footer())
	}
	l.keys("right")
	if !strings.Contains(footer(), "Enter open Health") {
		t.Fatalf("the footer follows the button: %q", footer())
	}
	l.keys("enter") // Open Health
	if l.sh().tab != commands.TabDoctor {
		t.Fatalf("Enter on Open Health showed %q", l.sh().tab)
	}
	l = newHomeLab(t, 160, 44)
	l.keys("enter") // Fix: the first button, as before
	if l.sh().fix == nil {
		t.Fatal("Enter on the first button did not open the fix")
	}
	l = newHomeLab(t, 160, 44)
	l.keys("end", "enter") // Open Deps
	if l.sh().tab != commands.TabDeps {
		t.Fatalf("Enter on the last item showed %q", l.sh().tab)
	}
}

func TestAClickOnAnItemSelectsItAndAClickOnItsButtonRunsIt(t *testing.T) {
	l := newHomeLab(t, 160, 44)
	l.clickText("Updates not looked at yet")
	if sh := l.sh(); sh.homeSel != 3 || sh.tab != commands.TabHome {
		t.Fatalf("the click selected %d on %q", sh.homeSel, sh.tab)
	}
	l.has("[Open Deps (Enter)]", "Needs you now  4/4")
	l.clickText("[Open Release")
	if l.sh().tab != commands.TabRelease {
		t.Fatalf("the button showed %q", l.sh().tab)
	}
}

func TestANarrowHomeShowsTheKeyOfEachItemAndAClickOnItRuns(t *testing.T) {
	l := newHomeLab(t, 80, 30)
	l.hasNot("[Fix (", "Start something")
	x, y := l.find("2 checks fail")
	line := []rune(l.lines()[y])
	if got := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(string(line[x:])), "│")); !strings.HasSuffix(got, "Enter") {
		t.Fatalf("the selected item ends in Enter: %q", got)
	}
	_, y = l.find("is affected by your changes")
	row := strings.TrimRight(strings.TrimSuffix(strings.TrimRight(l.lines()[y], " "), "│"), " ")
	if !strings.HasSuffix(row, " w") {
		t.Fatalf("the other items show their key: %q", row)
	}
	l.send(click(len([]rune(row))-1, y)).send(letGo(len([]rune(row))-1, y))
	if l.lastStarted() != "run lint test --affected" {
		t.Fatalf("the key started %q", l.lastStarted())
	}
}

func TestAClickOnTheLastRunBoxOpensItInTheRunPanel(t *testing.T) {
	l := newHomeLab(t, 160, 44)
	finishedRun(l, "run", "test")
	l.has("Last run", "last run just now")
	l.clickText("Last run")
	if l.sh().tab != commands.TabProjects || !l.sh().runShown() {
		t.Fatalf("tab %q, run shown %v", l.sh().tab, l.sh().runShown())
	}
}

func TestTheKeyboardTargetStaysInViewWhileHomeScrolls(t *testing.T) {
	l := newHomeLab(t, 120, 14)
	l.keys("end")
	l.has("[Open Deps (Enter)]", moreAbove)
	l.keys("home")
	l.has("[Fix (Enter)]")
	// Each step keeps both lines of the item on screen.
	for i := range homeItems(l) {
		if i > 0 {
			l.keys("down")
		}
		if !strings.Contains(l.view(), "(Enter)]") {
			t.Fatalf("after %d steps the item is out of view:\n%s", i, l.view())
		}
	}
}

func TestAClickOnAScrolledHomeHitsTheButtonUnderIt(t *testing.T) {
	l := showOnHome(newHomeLab(t, 120, 14), "[Open Release")
	l.has("[Open Release")
	l.clickText("[Open Release")
	if l.sh().tab != commands.TabRelease {
		t.Fatalf("the click showed %q", l.sh().tab)
	}
}

func TestTheFooterOffersScrollingOnlyWhenHomeScrolls(t *testing.T) {
	l := newHomeLab(t, 160, 44)
	l.has("move")
	l.hasNot("PgUp/PgDn")
	l = newHomeLab(t, 120, 14)
	l.has("move", "PgUp/PgDn scroll")
}

func TestHomeDrawsAtEveryHeightFromTheFloorUp(t *testing.T) {
	for _, w := range []int{72, 100, 160} {
		for h := minHeight; h <= 30; h++ {
			l := newHomeLab(t, w, h)
			lines := l.lines()
			if len(lines) != h {
				t.Fatalf("%dx%d: %d rows", w, h, len(lines))
			}
			// Page down until the end of Home shows: nothing stays out of reach.
			// A narrow Home ends with its last item, a wide one with Start something.
			last := "Updates not looked at yet"
			if l.sh().mainInner().W >= dashboard.WideFrom {
				last = "Compose a stack"
			}
			showOnHome(l, last)
			if !strings.Contains(l.view(), last) {
				t.Fatalf("%dx%d: the end of Home cannot be reached:\n%s", w, h, l.view())
			}
		}
	}
}
