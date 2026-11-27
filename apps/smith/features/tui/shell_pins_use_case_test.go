package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
)

func TestPIsPinsTheProjectToTheRailAndKeepsItForThePerson(t *testing.T) {
	l := newLab(t, true, 150, 44)
	var saved [][]string
	sh := l.sh()
	sh.savePins = func(p []string) { saved = append(saved, p) }
	l.m = sh
	l.openTab(commands.TabProjects)
	name := l.sh().list.SelectedName()
	l.keys("P")
	if len(saved) != 1 || !reflect.DeepEqual(saved[0], []string{name}) {
		t.Fatalf("the pin is kept: %v", saved)
	}
	rail := l.inside(l.sh().lay().Rail)
	text := strings.Join(rail, "\n")
	if !strings.Contains(text, "Pinned") || !strings.Contains(text, "– "+name) {
		t.Fatalf("the rail lists the pin under Pinned:\n%s", text)
	}
	l.keys("P")
	if len(saved) != 2 || len(saved[1]) != 0 {
		t.Fatalf("P again unpins it: %v", saved)
	}
	if strings.Contains(strings.Join(l.inside(l.sh().lay().Rail), "\n"), "Pinned") {
		t.Fatal("a rail without pins draws no heading")
	}
}

func TestAClickOnAPinOpensTheProject(t *testing.T) {
	l := newLab(t, true, 150, 44)
	sh := l.sh()
	sh.pins = []string{"kit"}
	l.m = sh
	l.openTab(commands.TabHome)
	l.keys("2", "/")
	l.typeText("web")
	l.keys("enter")
	l.openTab(commands.TabDeps)
	x, y := l.find("– kit")
	l.send(click(x+2, y)).send(letGo(x+2, y))
	if l.sh().tab != commands.TabProjects || l.sh().list.SelectedName() != "kit" {
		t.Fatalf("the click opens kit in Projects: %q, %q", l.sh().tab, l.sh().list.SelectedName())
	}
	if l.sh().list.Filter() != "" {
		t.Fatalf("a filter that hides the project goes: %q", l.sh().list.Filter())
	}
}

func TestThePaletteOffersThePinAndThenTheUnpin(t *testing.T) {
	l := newLab(t, true, 150, 44)
	l.openTab(commands.TabProjects)
	title := func() string {
		for _, e := range commands.PaletteEntries(l.sh().state()) {
			if e.ID == "projects.pin" {
				return e.Title
			}
		}
		return ""
	}
	if got := title(); got != "Pin the project to the rail" {
		t.Fatalf("pin: %q", got)
	}
	l.keys("P")
	if got := title(); got != "Unpin the project from the rail" {
		t.Fatalf("a pinned project offers the unpin: %q", got)
	}
}
