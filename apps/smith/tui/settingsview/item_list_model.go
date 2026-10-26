package settingsview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Section is one entry of the list on the left.
type Section int

// The sections, in the order of the list.
const (
	Plugins Section = iota
	Style
	Workspace
	Recipes
	Forms
)

// Sections lists them in order.
var Sections = []Section{Plugins, Style, Workspace, Recipes, Forms}

// Title is the name in the list.
func (s Section) Title() string {
	return [...]string{"Plugins", "Style", "Workspace", "Recipes", "Forms"}[s]
}

// About is the line under the heading of the open section.
func (s Section) About() string {
	return [...]string{
		"which plugins this workspace uses",
		"the colours and symbols of this interface, yours alone",
		"the settings of smith.toml",
		"the recipes of this workspace",
		"guided forms that run a smith command",
	}[s]
}

// Form is a guided form the Forms section opens.
type Form struct{ Title, Hint string }

type kind int

const (
	kindPlugin kind = iota
	kindExternal
	kindInstalled
	kindInstall
	kindTheme
	kindIcons
	kindSetting
	kindRecipe
	kindForm
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
			add(kindPlugin, "Built-in plugins", p.Name, i)
		}
		for i, p := range m.snap.External {
			add(kindExternal, "External plugins this workspace uses", p.Name, i)
		}
		for i, p := range m.snap.Installed {
			add(kindInstalled, "Installed on the search path", p.Name+"\x00"+p.Dir, i)
		}
		add(kindInstall, "Install", "install", 0)
	case Style:
		for i, t := range theme.ThemeNames {
			add(kindTheme, "Theme", t, i)
		}
		for i, t := range theme.IconSetNames {
			add(kindIcons, "Symbols", t, i)
		}
	case Workspace:
		for i, st := range m.snap.Settings {
			add(kindSetting, tableOf(st.Path), st.Path, i)
		}
	case Recipes:
		for i, r := range m.snap.Recipes {
			add(kindRecipe, "Recipe files", r.Path, i)
		}
	case Forms:
		for i, f := range m.forms {
			add(kindForm, "Guided forms", f.Title, i)
		}
	}
	return out
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
	for _, s := range Sections {
		m.cursor[s] = max(min(m.cursor[s], len(m.items(s))-1), 0)
	}
	return m
}
