package tui

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/projectpage"
)

func TestSlashTypesTheFilterAndTheListNarrowsAtOnce(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("/")
	if !l.sh().filtering {
		t.Fatal("/ did not give the keyboard to the filter")
	}
	l.has("Filter: ▏", "Enter ")
	l.typeText("web")
	if len(l.started) != 0 {
		t.Fatalf("typing ran a key: %v", l.started) // b is the build key outside the filter
	}
	listed, _ := l.sh().list.Listed()
	if len(listed) != 1 || listed[0].Name != "web" || l.sh().selectedProject() != "web" {
		t.Fatalf("the filter lists %v and selects %q", listed, l.sh().selectedProject())
	}
	if top := mainTop(l); !strings.Contains(top, "─ Projects 1/4 ") {
		t.Fatalf("the title counts the listed projects: %q", top)
	}
	l.has("Filter: web▏", "0 of 1 lib")
	l.hasNot("–  kit")
	l.keys("enter")
	if l.sh().filtering || l.sh().list.Filter() != "web" {
		t.Fatalf("Enter keeps the filter and ends the typing: typing %v, filter %q", l.sh().filtering, l.sh().list.Filter())
	}
	l.has("Filter: web")
	l.hasNot("Filter: web▏")
	l.keys("esc")
	if l.sh().list.Filter() != "" {
		t.Fatal("Esc on the list clears the filter")
	}
	l.has("–  kit", "┬─ Projects ")
}

func TestEscWhileTypingClearsTheFilterAndBackspaceRubsOut(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("/").typeText("kind:lib")
	l.keys("backspace", "backspace", "backspace")
	if got := l.sh().list.Filter(); got != "kind:" {
		t.Fatalf("backspace left %q", got)
	}
	l.keys("esc")
	if l.sh().filtering || l.sh().list.Filter() != "" {
		t.Fatalf("Esc ends the typing and clears the filter: typing %v, filter %q", l.sh().filtering, l.sh().list.Filter())
	}
}

func TestAFilterThatListsNothingSaysHowToClearIt(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("/").typeText("zzz").keys("enter")
	l.has("No project matches the filter.", "clears it.") // the line wraps in the list
	l.hasNot("Last result")                               // no project is selected, so the detail tells the workspace
}

func TestTheFilterStaysWhenTheProjectsAreReadAgain(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("/").typeText("kind:app").keys("enter")
	sh := l.sh()
	l.drive(sh.loadProjects())
	if got := l.sh().list.Filter(); got != "kind:app" {
		t.Fatalf("a reload dropped the filter: %q", got)
	}
}

func TestOnANarrowScreenEnterOpensTheSectionsAndEscGoesBack(t *testing.T) {
	l := newLab(t, true, 80, 30)
	l.keys("enter")
	if !l.sh().detailOpen || !l.sh().projectKeys() {
		t.Fatalf("Enter opens the detail with its keys: open %v, keys %v", l.sh().detailOpen, l.sh().projectKeys())
	}
	if top := mainTop(l); !strings.HasPrefix(top, "╭─ ‹ Projects / orders ") {
		t.Fatalf("the frame names the way back: %q", top)
	}
	l.has("Tasks   Deps   Runs   Settings", "Tab section", "Enter run", "Esc back")
	l.keys("tab")
	if l.sh().projSection != projectpage.Dependencies || !l.sh().detailOpen {
		t.Fatalf("Tab shows the next section in the open detail: section %d, open %v", l.sh().projSection, l.sh().detailOpen)
	}
	l.has("Uses", "Used by")
	l.keys("esc")
	if l.sh().detailOpen {
		t.Fatal("Esc did not go back to the list")
	}
	l.has("Filter:", "APPS")
}

func TestTheSettingsSectionIsTheEditorOfTheProjectAndTheSettingsTabNoLongerHasIt(t *testing.T) {
	l, _ := newSettingsLab(t, 140, 40)
	l.openTab(commands.TabProjects)
	l.keys("R")
	l.runDetailAction("[Settings…")
	if l.sh().projSection != projectpage.Settings || l.sh().projMenu {
		t.Fatalf("Settings… shows the Settings section: section %d, menu %v", l.sh().projSection, l.sh().projMenu)
	}
	if l.sh().settings.ProjectOpen() != "orders" {
		t.Fatalf("the section reads the settings of orders, open %q", l.sh().settings.ProjectOpen())
	}
	l.has("Project orders")
	// The letters of the section are its own: b does not build here.
	l.keys("b")
	if len(l.started) != 0 {
		t.Fatalf("b in the Settings section started %v", l.started)
	}
	// Another project in the list: the section follows it.
	l.keys("esc", "down")
	l.has("Project web")
	// The Settings tab lists its own sections only.
	x, y := l.railSpot(commands.TabSettings)
	l.sendPatiently(click(x+1, y))
	if l.sh().tab != commands.TabSettings || l.sh().settings.ProjectOpen() != "" {
		t.Fatalf("the Settings tab keeps the project %q", l.sh().settings.ProjectOpen())
	}
	l.hasNot("Project web")
}

func TestThePaletteOfProjectsListsTheActionsOfTheSelectedProject(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("down", "up") // the palette is rebuilt after each event, on the selected orders
	var ids, titles []string
	for _, e := range l.sh().pal.Entries() {
		ids, titles = append(ids, e.ID), append(titles, e.Title)
	}
	for _, id := range []string{"x:explain", "x:why", "x:graph"} {
		if !slices.Contains(ids, id) {
			t.Fatalf("the palette lacks %s: %v", id, ids)
		}
	}
	if slices.Contains(ids, "x:settings") || !slices.Contains(titles, "How was this resolved? (orders)") {
		t.Fatalf("entries %v", titles)
	}
	home := newHomeLab(t, 140, 34)
	for _, e := range home.sh().pal.Entries() {
		if strings.HasPrefix(e.ID, "x:") {
			t.Fatalf("Home lists an action of a project: %s", e.ID)
		}
	}
	l.keys("ctrl+k").typeText("graph").keys("enter")
	if got := l.lastStarted(); got != "graph --focus orders" && got != "graph" {
		t.Fatalf("the palette entry started %q", got)
	}
}

func TestTheOpenButtonOpensTheFolderOfTheProject(t *testing.T) {
	l := newLab(t, true, 140, 34)
	sh := l.sh()
	sh.wsRoot = filepath.Join("ws", "acme")
	l.m = sh
	l.clickText("[Open]")
	if want := filepath.Join("ws", "acme", "apps", "orders"); len(l.opened) != 1 || filepath.Clean(l.opened[0]) != want {
		t.Fatalf("opened %v, want %s", l.opened, want)
	}
	if !strings.Contains(l.sh().hint, "Opened") {
		t.Fatalf("the status line says what opened: %q", l.sh().hint)
	}
}

func TestWithoutAWayToOpenFilesTheDetailHasNoOpenButton(t *testing.T) {
	l := newLab(t, true, 140, 34)
	sh := l.sh()
	sh.sources.Open = nil
	l.m = sh
	l.has("Run… (R)")
	l.hasNot("[Open]")
}
