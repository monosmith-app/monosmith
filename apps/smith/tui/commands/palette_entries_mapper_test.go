package commands

import (
	"strings"
	"testing"
)

func TestThePaletteListsEachIntentOnceWithAUniqueKeyInEveryView(t *testing.T) {
	for _, view := range []string{TabHome, TabProjects, TabRun, TabDoctor, TabDeps, TabRelease, TabSettings} {
		s := State{View: view, InWorkspace: true, Project: "web", Projects: 4}
		titles, keys := map[string]bool{}, map[string]string{}
		for _, e := range PaletteEntries(s) {
			if titles[strings.ToLower(e.Title)] {
				t.Errorf("%s: %q is listed twice", view, e.Title)
			}
			titles[strings.ToLower(e.Title)] = true
			if e.Shortcut == "" || e.Section == "" {
				continue
			}
			if other, ok := keys[e.Shortcut]; ok && e.Shortcut != "1" {
				t.Errorf("%s: %q and %q share the key %s", view, other, e.Title, e.Shortcut)
			}
			keys[e.Shortcut] = e.Title
		}
	}
}

func TestThePaletteNamesTheScopeOfARunAndTheCommandBehindIt(t *testing.T) {
	var build, all, install PaletteEntry
	for _, e := range PaletteEntries(State{View: TabHome, InWorkspace: true, Project: "web", Projects: 23}) {
		switch e.ID {
		case "run.build":
			build = e
		case "run.build-all":
			all = e
		case "projects.install-package":
			install = e
		}
	}
	if build.Title != "Build web" || build.Hint != "smith run build -p web" {
		t.Errorf("build is %+v", build)
	}
	if all.Title != "Build all 23" || all.Hint != "smith run build" || all.Shortcut != "B" {
		t.Errorf("build all is %+v", all)
	}
	if install.Title != "Install a package…" || install.Unavailable != "" {
		t.Errorf("install a package is %+v", install)
	}
}

func TestThePaletteHasNoRawCommandNamesAndNoFreeTextInstall(t *testing.T) {
	for _, e := range PaletteEntries(State{View: TabHome, InWorkspace: true}) {
		if e.Title == strings.ToLower(e.Title) {
			t.Errorf("%q reads like a raw command name", e.Title)
		}
		if e.Title == "install" || e.Title == "uninstall" || e.Title == "info" {
			t.Errorf("%q is a raw command", e.Title)
		}
	}
}

func TestTheFooterNamesTheScopeOfTheRunKeys(t *testing.T) {
	view, _ := Hints(State{View: TabHome, InWorkspace: true, Project: "markobot-vscode-extension", Projects: 23})
	var labels []string
	for _, h := range view {
		labels = append(labels, h.Glyph+" "+h.Label)
	}
	got := strings.Join(labels, "|")
	for _, want := range []string{"b build markobot-vscode…", "B build all 23", "t test markobot-vscode…"} {
		if !strings.Contains(got, want) {
			t.Errorf("the footer lacks %q: %s", want, got)
		}
	}
}
