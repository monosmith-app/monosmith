package settingsview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Section is one entry of the list on the left.
type Section int

// The sections, in the order of the list.
const (
	Plugins Section = iota
	Architecture
	Style
	Workspace
	Recipes
	Assistants
	CI
	Forms
	Update
	// Project is the settings of one project, listed last while one is open
	// (see Model.OpenProject).
	Project
)

// sectionCount is the number of sections, the size of the per-section state.
const sectionCount = Project + 1

// Sections lists the sections that are always there, in order; the Project
// section follows them while a project is open.
var Sections = []Section{Plugins, Architecture, Style, Workspace, Recipes, Assistants, CI, Forms, Update}

// sectionForms are the actions (by catalogue id, see guide.Action.Presents)
// whose guided forms a section lists besides its own rows. The Forms section
// lists every form.
var sectionForms = map[Section][]string{
	Plugins:    {"plugin set"},                   // words:literal
	Assistants: {"mcp install", "agent install"}, // words:literal
	CI:         {"ci", "pipeline"},
}

// Title is the name in the list.
func (s Section) Title() string {
	return [...]string{words.Say("settingsview.section.plugins"), words.Say("settingsview.section.architecture"), words.Say("settingsview.section.style"), words.Say("settingsview.section.workspace"), words.Say("settingsview.section.recipes"), words.Say("settingsview.section.assistants"), words.Say("settingsview.section.ci"), words.Say("settingsview.section.forms"), words.Say("settingsview.section.update"), words.Say("settingsview.section.project")}[s]
}

// About is the line under the heading of the open section.
func (s Section) About() string {
	return [...]string{
		words.Say("settingsview.about.plugins"),
		words.Say("settingsview.about.architecture"),
		words.Say("settingsview.about.style"),
		words.Say("settingsview.about.workspace"),
		words.Say("settingsview.about.recipes"),
		words.Say("settingsview.about.assistants"),
		words.Say("settingsview.about.ci"),
		words.Say("settingsview.about.forms"),
		words.Say("settingsview.about.update"),
		words.Say("settingsview.about.project"),
	}[s]
}

// Form is a guided form the Forms section opens. Presents is the id of the
// action it stands for ("mcp install"), which places it in a section too.
type Form struct{ Title, Hint, Presents string }

// formsOf are the forms a section lists besides its own rows, in the order of
// sectionForms; a form the workspace does not offer is left out.
func (m Model) formsOf(s Section) []int {
	var out []int
	for _, id := range sectionForms[s] {
		for i, f := range m.forms {
			if f.Presents == id {
				out = append(out, i)
				break
			}
		}
	}
	return out
}

type kind int

const (
	kindPlugin kind = iota
	kindExternal
	kindInstalled
	kindInstall
	kindSearch
	kindArchitecture
	kindTheme
	kindIcons
	kindSetting
	kindRecipe
	kindForm
	kindUpdate
	kindUpdateMute
	// kindFact is a value of a project that its manifest owns: shown, not changed.
	kindFact
)

// item is one selectable row of a section; group is the heading it sits under.
type item struct {
	kind  kind
	group string
	// key names it: a plugin, a setting's path, a choice, a form's title.
	key string
	// idx is its place in the snapshot's slice (or the choice list).
	idx int
}

// items are the selectable rows of a section, in the order they are drawn.
func (m Model) items(s Section) []item {
	var out []item
	add := func(k kind, group, key string, idx int) { out = append(out, item{k, group, key, idx}) }
	switch s {
	case Plugins:
		for i, p := range m.snap.Plugins {
			add(kindPlugin, words.Say("settingsview.group.builtin"), p.Name, i)
		}
		for i, p := range m.snap.External {
			add(kindExternal, words.Say("settingsview.group.external"), p.Name, i)
		}
		for i, p := range m.snap.Installed {
			add(kindInstalled, words.Say("settingsview.group.installed"), p.Name+"\x00"+p.Dir, i)
		}
		add(kindSearch, words.Say("settingsview.group.install"), "search", 0)
		add(kindInstall, words.Say("settingsview.group.install"), "install", 0)
	case Architecture:
		for i, a := range m.snap.Architectures {
			add(kindArchitecture, words.Say("settingsview.group.layout"), a.Name, i)
		}
	case Style:
		for i, t := range theme.ThemeNames {
			add(kindTheme, words.Say("settingsview.group.theme"), t, i)
		}
		for i, t := range theme.IconSetNames {
			add(kindIcons, words.Say("settingsview.group.symbols"), t, i)
		}
	case Workspace:
		for i, st := range m.snap.Settings {
			add(kindSetting, tableOf(st.Path), st.Path, i)
		}
	case Project:
		for i, f := range m.proj.Facts {
			add(kindFact, words.Say("settingsview.group.facts", "file", f.Source), f.Key, i)
		}
		for i, st := range m.proj.Settings {
			add(kindSetting, words.Say("settingsview.group.project", "table", m.projectTable()), st.Path, i)
		}
	case Recipes:
		for i, r := range m.snap.Recipes {
			add(kindRecipe, words.Say("settingsview.group.recipes"), r.Path, i)
		}
	case Update:
		add(kindUpdate, words.Say("settingsview.group.update"), "update", 0)
		add(kindUpdateMute, words.Say("settingsview.group.update"), "mute", 0)
	case Forms:
		for i, f := range m.forms {
			add(kindForm, words.Say("settingsview.group.forms"), f.Title, i)
		}
	}
	for _, i := range m.formsOf(s) {
		add(kindForm, words.Say("settingsview.group.forms"), m.forms[i].Title, i)
	}
	return out
}

// readsWorkspace is false for a section whose rows need no snapshot: the
// style, which is the person's, and the sections that only list forms.
func (s Section) readsWorkspace() bool {
	return s != Style && s != Forms && s != Assistants && s != CI && s != Update
}

// tableOf is the smith.toml table a setting path belongs to.
func tableOf(path string) string {
	table, _, _ := strings.Cut(path, ".")
	return "[" + table + "]"
}

// current is the item under the cursor of the open section.
func (m Model) current() (item, bool) {
	list := m.items(m.section)
	c := m.cursor[m.section]
	if c < 0 || c >= len(list) {
		return item{}, false
	}
	return list[c], true
}

// clampCursors keeps every cursor on a row, after the lists changed.
func (m Model) clampCursors() Model {
	for _, s := range m.sections() {
		m.cursor[s] = max(min(m.cursor[s], len(m.items(s))-1), 0)
	}
	return m
}

// sections are the sections the list shows: Sections, and Project while a
// project is open.
func (m Model) sections() []Section {
	if m.project == nil {
		return Sections
	}
	return append(append([]Section{}, Sections...), Project)
}

// settingAt is the setting of a kindSetting row: the open project's in the
// Project section, the workspace's elsewhere.
func (m Model) settingAt(it item) Setting {
	if m.section == Project {
		return m.proj.Settings[it.idx]
	}
	return m.snap.Settings[it.idx]
}
