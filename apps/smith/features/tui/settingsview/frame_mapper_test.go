package settingsview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

func paneText(lines []string) string { return pane.Strip(strings.Join(lines, "\n")) }

func TestTheSectionsPaneMarksTheSelectedSectionAndSaysWhereProjectSettingsAre(t *testing.T) {
	m := opened(t, sample()).Show(Workspace).ToSections()
	out := paneText(m.SectionsPane(plainTheme(), 34, 36, -1, -1))
	for _, want := range []string{"▶ Workspace", "  Plugins", "  Update", "Per-project settings are in", "Projects › a project › Settings"} {
		if !strings.Contains(out, want) {
			t.Fatalf("the sections lack %q:\n%s", want, out)
		}
	}
	if !m.OnSections() {
		t.Fatal("ToSections puts the keyboard on the sections")
	}
}

func TestTheFieldsPaneIsATableWithTheBoxOfTheSelectedSettingAndItsButtons(t *testing.T) {
	m := opened(t, sample()).Show(Workspace)
	lines := m.FieldsPane(plainTheme(), 90, 36, -1, -1)
	out := paneText(lines)
	for _, want := range []string{
		"Setting", "Value", "Default", "Type",
		"ci.devcontainer", "○ off", "switch", "ci.providers", "ticks", "cache.max_size_mb", "512", "number",
		"╭─ ci.devcontainer ", "Write a devcontainer.", "$ smith config set ci.devcontainer false",
		"Writes smith.toml exactly as the command does.", "[Save (Enter)]", "[Reset to default (u)]",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("the fields lack %q:\n%s", want, out)
		}
	}
	if m.PaneTitle() != "Workspace · smith.toml" || m.Show(Style).PaneTitle() != "Style" {
		t.Fatalf("titles %q, %q", m.PaneTitle(), m.Show(Style).PaneTitle())
	}
	if len(lines) != 36 {
		t.Fatalf("%d lines", len(lines))
	}
}

func TestSaveOpensTheChoiceAndResetUnsetsOnlyAWrittenValue(t *testing.T) {
	f := sample()
	m := opened(t, f).Show(Workspace)
	w, h := 90, 36
	plain := plainTheme()
	row, x := -1, -1
	for y, l := range m.FieldsPane(plain, w, h, -1, -1) {
		if i := strings.Index(pane.Strip(l), "[Save"); i >= 0 {
			row, x = y, i+1
		}
	}
	if row < 0 {
		t.Fatal("no Save button")
	}
	m, res := m.FieldsClick(w, h, x, row, false)
	if _, more := settle(m, res); !m.Modal() || len(f.calls) != 0 || len(more) != 0 {
		t.Fatalf("Save on a switch presents its choice and writes nothing: %v", f.calls)
	}
	m, _ = press(m, "esc", "down") // ci.providers, written
	reset := strings.Index(pane.Strip(m.FieldsPane(plain, w, h, -1, -1)[row]), "[Reset") + 1
	m, res = m.FieldsClick(w, h, reset, row, false)
	settle(m, res)
	if got := strings.Join(f.calls, ";"); got != "unset ci.providers" {
		t.Fatalf("Reset unsets the written value: %q", got)
	}
}

func TestAConfirmedSwitchWritesThroughTheBackendAndNotDecidedUnsetsIt(t *testing.T) {
	f := sample()
	f.snap.Settings = []Setting{{Path: "release.nuget_trusted_publishing", Type: "boolean", Explanation: "Publish with GitHub OIDC.", Value: "true", Set: true, Current: []string{"true"}}}
	m := opened(t, f).Show(Workspace)
	m, _ = press(m, "enter")
	out := text(m, 100, 30)
	for _, want := range []string{"( ) Not decided", "(●) On", "( ) Off", "$ smith config set release.nuget_trusted_publishing true"} {
		if !strings.Contains(out, want) && !strings.Contains(out, strings.Replace(want, "●", "*", 1)) {
			t.Fatalf("the choice lacks %q:\n%s", want, out)
		}
	}
	m, _ = press(m, "down")
	if out := text(m, 100, 30); !strings.Contains(out, "$ smith config set release.nuget_trusted_publishing false") {
		t.Fatalf("the command line follows the cursor:\n%s", out)
	}
	if len(f.calls) != 0 {
		t.Fatalf("nothing is written before Enter: %v", f.calls)
	}
	m, _ = press(m, "up", "up")
	if out := text(m, 100, 30); !strings.Contains(out, "$ smith config unset release.nuget_trusted_publishing") {
		t.Fatalf("Not decided is an unset:\n%s", out)
	}
	m, _ = press(m, "enter")
	if got := strings.Join(f.calls, ";"); got != "unset release.nuget_trusted_publishing" || m.Modal() {
		t.Fatalf("calls %q, modal %v", got, m.Modal())
	}
}

func TestNarrowShowsTheSectionsThenEnterOpensOneAndEscGoesBack(t *testing.T) {
	m := opened(t, sample()).ToSections()
	out := paneText(m.Narrow(plainTheme(), 78, 25, -1, -1))
	for _, want := range []string{"▶ Plugins", "2 on · 1 external", "Workspace", "5 settings, 2 set", "2 recipes", "theme auto · symbols auto"} {
		if !strings.Contains(out, want) {
			t.Fatalf("the narrow sections lack %q:\n%s", want, out)
		}
	}
	m, _ = press(m, "down", "down", "down", "enter")
	if m.OnSections() || m.Section() != Workspace {
		t.Fatalf("Enter opens the section: on sections %v, %v", m.OnSections(), m.Section())
	}
	if out := paneText(m.Narrow(plainTheme(), 78, 25, -1, -1)); !strings.Contains(out, "ci.devcontainer") || strings.Contains(out, "Per-project") {
		t.Fatalf("the open section fills the pane:\n%s", out)
	}
	m, _ = press(m, "enter")
	if !m.Modal() {
		t.Fatal("Enter on a field edits it")
	}
	m, _ = press(m, "esc")
	if m.Modal() || m.OnSections() {
		t.Fatal("the first Esc closes the editor")
	}
	m, _ = press(m, "esc")
	if !m.OnSections() {
		t.Fatal("the second Esc goes back to the sections")
	}
}

func TestANarrowClickSelectsASectionAndASecondClickOpensIt(t *testing.T) {
	m := opened(t, sample()).ToSections()
	w, h := 78, 25
	row := -1
	for y, l := range m.Narrow(plainTheme(), w, h, -1, -1) {
		if strings.Contains(pane.Strip(l), "Recipes") {
			row = y
		}
	}
	m, _ = m.NarrowClick(w, h, 4, row, false)
	if m.Section() != Recipes || !m.OnSections() {
		t.Fatalf("one click selects: %v, on sections %v", m.Section(), m.OnSections())
	}
	m, _ = m.NarrowClick(w, h, 4, row, false)
	if m.OnSections() {
		t.Fatal("a click on the selected section opens it")
	}
	if m = m.ToSections().SectionsClick(34, 36, 1, true); m.OnSections() || m.Section() != Plugins {
		t.Fatalf("a double click in the wide list opens Plugins: %v", m.Section())
	}
}

func TestThePanesFitTheirAreaAtTheScreenSizes(t *testing.T) {
	for _, sz := range [][2]int{{36, 36}, {89, 36}, {78, 24}, {59, 32}, {78, 25}} {
		for _, sec := range Sections {
			for _, onItems := range []bool{true, false} {
				m := opened(t, sample()).Show(sec)
				if !onItems {
					m = m.ToSections()
				}
				for name, lines := range map[string][]string{
					"sections": m.SectionsPane(plainTheme(), sz[0], sz[1], -1, -1),
					"fields":   m.FieldsPane(plainTheme(), sz[0], sz[1], -1, -1),
					"narrow":   m.Narrow(plainTheme(), sz[0], sz[1], -1, -1),
				} {
					if len(lines) != sz[1] {
						t.Fatalf("%s %v %s: %d lines", name, sz, sec.Title(), len(lines))
					}
					for i, l := range lines {
						if pane.Width(l) != sz[0] {
							t.Fatalf("%s %v %s: line %d is %d cells: %q", name, sz, sec.Title(), i, pane.Width(l), l)
						}
					}
				}
			}
		}
	}
}
