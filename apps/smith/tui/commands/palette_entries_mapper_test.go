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
	for _, e := range PaletteEntries(State{View: TabProjects, InWorkspace: true, Project: "web", Projects: 23}) {
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
	view, _ := Hints(State{View: TabProjects, InWorkspace: true, Project: "markobot-vscode-extension", Projects: 23})
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

func TestThePaletteOffersTheTwoPointZeroCommands(t *testing.T) {
	want := map[string]string{
		"ws.undo": "Undo the last change…", "ws.report": "Open the report", "ws.agent-file": "Write the agent file…",
		"ws.derive-check": "Check the derived files", "ws.derive": "Update the derived files…", "ws.why": "Why is this file here?…",
	}
	for _, e := range PaletteEntries(State{View: TabHome, InWorkspace: true}) {
		if title, ok := want[e.ID]; ok {
			if e.Title != title || e.Unavailable != "" {
				t.Errorf("%s: %q unavailable %q", e.ID, e.Title, e.Unavailable)
			}
			delete(want, e.ID)
		}
	}
	for id := range want {
		t.Errorf("%s is not in the palette", id)
	}
	outside := map[string]bool{}
	for _, e := range PaletteEntries(State{View: TabHome}) {
		outside[e.ID] = true
	}
	for _, id := range []string{"ws.undo", "ws.report", "ws.agent-file", "ws.derive-check", "ws.derive", "ws.why"} {
		if outside[id] {
			t.Errorf("%s needs a workspace", id)
		}
	}
}

func TestDescribeSaysWhatTheNewCommandsChange(t *testing.T) {
	for argv, want := range map[string]string{
		"undo --yes": "Undo the last change", "agent file": "Write the agent file", "derive": "Update the derived files",
		"derive --eject a.json": "Own a derived file", "derive --adopt a.json": "Keep your edit", "derive --replace a.json": "Replace a derived file",
	} {
		if title, why := Describe(strings.Fields(argv)); !strings.HasPrefix(title, want) || why == "" {
			t.Errorf("%q: %q / %q", argv, title, why)
		}
	}
}
