package commands

import (
	"strings"
	"testing"
)

func TestABareInstallSaysItInstallsEverythingAndPointsToTheSinglePackageForm(t *testing.T) {
	title, why := Describe([]string{"install"})
	if title != "Install everything" {
		t.Fatalf("title %q", title)
	}
	if !strings.Contains(why, "Install a package") {
		t.Fatalf("it should say where one package is added: %q", why)
	}
	entry := func(id string) PaletteEntry {
		for _, e := range PaletteEntries(State{View: TabHome, InWorkspace: true}) {
			if e.ID == id {
				return e
			}
		}
		t.Fatalf("no palette entry %s", id)
		return PaletteEntry{}
	}
	if e := entry("proj.install"); e.Title != "Install everything…" || e.Hint != "smith install" {
		t.Fatalf("the whole-workspace entry: %+v", e)
	}
	if e := entry("projects.install-package"); e.Title != "Install a package…" {
		t.Fatalf("the single-package entry: %+v", e)
	}
}
