package settingsview

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// themeAbout and iconsAbout say what each choice of the Style section means.
var (
	themeAbout = map[string]string{
		theme.Auto: "follow the terminal's background",
		"dark":     "for a dark background",
		"light":    "for a light background",
		"plain":    "no colour: bold and reverse carry the meaning",
	}
	iconsAbout = map[string]string{
		theme.Auto: "unicode, unless the terminal cannot draw it",
		"unicode":  "✔ ✖ ▲ ● ⚡ and a braille spinner",
		"ascii":    "[v] [x] [!] and a plain spinner, for terminals without unicode",
	}
)

// itemRow is one row of a list.
func (m Model) itemRow(th theme.Theme, it item, w int) string {
	switch it.kind {
	case kindPlugin:
		p := m.snap.Plugins[it.idx]
		box, state := "[ ]", th.Muted("off")
		if p.Enabled {
			box, state = "[x]", th.Success("on")
		}
		return "   " + box + " " + pane.Cell(p.Name, 14) + state
	case kindExternal:
		p := m.snap.External[it.idx]
		mark, state := th.Success(th.Icons.Passed), "trusted"
		switch {
		case p.Problem != "":
			mark, state = th.Failure(th.Icons.Failed), "cannot start: "+p.Problem
		case !p.Trusted:
			mark, state = th.Warning(th.Icons.Warning), "not trusted"
		case p.Pinned:
			state += ", pinned"
		}
		return "   " + mark + " " + pane.Cell(p.Name, 14) + pane.Cell(state, 24) + th.Muted(p.Source)
	case kindInstalled:
		p := m.snap.Installed[it.idx]
		if p.Problem != "" {
			return "   " + th.Failure(th.Icons.Failed) + " " + pane.Cell(p.Name, 14) + th.Failure("invalid: "+p.Problem)
		}
		return "     " + pane.Cell(p.Name, 14) + pane.Cell(p.Version, 10) + th.Muted(p.Dir)
	case kindInstall:
		row := "   " + th.Accent("+ Install a plugin…")
		if m.b.Install == nil {
			row += "  " + th.Muted("smith plugin install <name|path>")
		}
		return row
	case kindArchitecture:
		a := m.snap.Architectures[it.idx]
		return "   " + m.radio(th, a.Name == m.snap.Architecture) + " " + pane.Cell(a.Name, 16) + th.Muted(a.Meaning)
	case kindTheme:
		return "   " + m.radio(th, normalise(it.key) == m.savedTheme) + " " + pane.Cell(it.key, 8) + th.Muted(themeAbout[it.key])
	case kindIcons:
		return "   " + m.radio(th, normalise(it.key) == m.savedIcons) + " " + pane.Cell(it.key, 8) + th.Muted(iconsAbout[it.key])
	case kindSetting:
		st := m.snap.Settings[it.idx]
		return "   " + pane.Cell(st.Path, m.pathWidth(w)) + shownValue(th, st)
	case kindRecipe:
		r := m.snap.Recipes[it.idx]
		if r.Error != "" {
			return "   " + th.Failure(th.Icons.Failed) + " " + pane.Cell(r.Path, 28) + th.Failure("invalid")
		}
		return "   " + th.Success(th.Icons.Passed) + " " + pane.Cell(r.Path, 28) + th.Muted(strings.TrimSpace(r.Name+" "+r.Version))
	case kindForm:
		f := m.forms[it.idx]
		return "   " + pane.Cell(f.Title, 28) + th.Muted(f.Hint)
	}
	return ""
}

// radio is the mark of a choice that is the saved one.
func (m Model) radio(th theme.Theme, on bool) string {
	if on {
		return "(" + th.Success(th.Icons.Dot) + ")"
	}
	return "( )"
}

// pathWidth is the width of the setting names: the longest, up to half the pane.
func (m Model) pathWidth(w int) int {
	n := 0
	for _, s := range m.snap.Settings {
		n = max(n, pane.Width(s.Path))
	}
	return min(n+2, max(w/2, 12))
}

// summary is one line about a row, for when there is no room for the detail.
func (m Model) summary(it item) string {
	switch it.kind {
	case kindPlugin:
		return "Space turns " + it.key + " on or off"
	case kindExternal:
		return "Enter reviews it before you trust it"
	case kindSetting:
		return oneLine(m.snap.Settings[it.idx].Explanation)
	case kindTheme, kindIcons:
		return "Enter applies it at once"
	case kindArchitecture:
		return "Enter chooses it and saves it in smith.toml"
	}
	return ""
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// emptyList is what a section with no rows says.
func (m Model) emptyList(th theme.Theme, w int) []string {
	if m.section == Recipes {
		return append([]string{""}, indent(pane.Wrap(th.Muted("No recipe files (*.recipe.toml) in this workspace. A recipe is a team's choice of plugins and settings; smith recipe check <file> checks one."), w-2, 0))...)
	}
	return []string{"", " " + th.Muted("Nothing here.")}
}

func (m Model) emptyDetail(th theme.Theme, w int) []string { return nil }

// detail tells about the row: what it is, what it does, which key changes it.
func (m Model) detail(th theme.Theme, it item, w int) []string {
	wrap := func(s string) []string { return pane.Wrap(s, max(w, 4), 0) }
	var out []string
	add := func(lines ...string) {
		for _, l := range lines {
			out = append(out, wrap(l)...)
		}
	}
	switch it.kind {
	case kindPlugin:
		add(th.Bold(it.key)+th.Muted("  support for a language, built into smith"),
			"Space turns it on or off for this workspace; at least one must stay on. Projects in a language that is off are not built, tested or checked.",
			th.Muted("Tools that are always on: "+strings.Join(m.snap.Addons, ", ")+". The same as: smith plugin enable|disable "+it.key))
	case kindExternal:
		p := m.snap.External[it.idx]
		add(th.Bold(p.Name)+th.Muted("  from "+p.Source), th.Muted("runs   ")+p.Command)
		sees := "the baseline environment only"
		if len(p.Env) > 0 {
			sees = "the baseline plus " + strings.Join(p.Env, ", ")
		}
		add(th.Muted("sees   ") + sees)
		switch {
		case p.Problem != "":
			add(th.Failure(p.Problem))
		case p.Trusted:
			add(th.Success("Trusted") + ". Enter reviews it again, f forgets the decision (smith plugin untrust).")
		default:
			add(th.Warning("Not trusted") + ": smith will not start it. Enter shows what it is, then trusts it (smith plugin trust). " + p.Detail)
		}
	case kindInstalled:
		p := m.snap.Installed[it.idx]
		add(th.Bold(p.Name)+th.Muted("  "+p.Version), th.Muted(p.Dir))
		if p.Problem != "" {
			add(th.Failure(p.Problem))
		} else {
			add("A workspace starts it when smith.toml declares it under [plugins.external." + p.Name + "].")
		}
		add(th.Muted("Search path: " + strings.Join(m.snap.SearchPath, ", ")))
	case kindInstall:
		add(th.Bold("Install a plugin"), "From a name or a path: smith plugin install <name|path>.")
		if m.b.Install == nil {
			add(th.Muted("This smith does not have that command yet. When it does, Enter asks for the name here and installs it without leaving."))
		} else {
			add("Enter asks for the name or path and installs it.")
		}
	case kindArchitecture:
		a := m.snap.Architectures[it.idx]
		add(th.Bold(a.Name)+th.Muted("  "+a.Meaning),
			"The architecture is how the code of this workspace is laid out. smith checks that the code follows the one you choose, and the check fails when it does not.",
			"Enter chooses it and saves it at once: it "+a.Writes+". Choose none to stop the checks.",
			th.Muted("The same as: smith architecture "+a.Name))
	case kindTheme:
		add(m.previewLines(th, it, w)...)
	case kindIcons:
		add(m.previewLines(th, it, w)...)
	case kindSetting:
		st := m.snap.Settings[it.idx]
		add(th.Bold(st.Path)+th.Muted("  "+st.Type), st.Explanation)
		if st.Default != "" {
			add(th.Muted("Default: ") + st.Default)
		}
		if len(st.Choices) > 0 {
			add(th.Muted("Choices: ") + strings.Join(st.Choices, ", "))
		}
		now := shownValue(th, st)
		if !st.Set {
			now += th.Muted(" (the default)")
		}
		keys := "Enter changes it"
		if st.Set {
			keys += ", u resets it to its default"
		}
		add(th.Muted("Now: ")+now, th.Muted(keys+" (smith config set|unset "+st.Path+")"))
	case kindRecipe:
		r := m.snap.Recipes[it.idx]
		add(th.Bold(r.Path))
		if r.Error != "" {
			add(th.Failure(r.Error))
		} else {
			add(strings.TrimSpace(r.Name + " " + r.Version))
			if r.Extends != "" {
				add(th.Muted("extends ") + r.Extends)
			}
		}
		add(th.Muted("smith recipe check " + r.Path + " shows what it would use and where each value came from."))
	case kindForm:
		f := m.forms[it.idx]
		add(th.Bold(f.Title), f.Hint, th.Muted("Enter opens the form; it ends as a smith command line."))
	}
	return out
}

// editorLines draws the open editor in w by h cells.
func (m Model) editorLines(th theme.Theme, w, h int) []string {
	e := m.ed
	var lines []string
	add := func(s string) { lines = append(lines, pane.Wrap(" "+s, max(w, 4), 1)...) }
	lines = append(lines, "")
	add(th.Bold(e.title))
	if e.about != "" {
		add(th.Muted(e.about))
	}
	var keys string
	switch e.kind {
	case editText, editInstall:
		if e.setting.Default != "" {
			add(th.Muted("Default: " + e.setting.Default))
		}
		if e.setting.Type == "array" {
			add(th.Muted("Separate the items with commas."))
		}
		in := e.input
		in.Width = max(w-6, 4)
		lines = append(lines, "", " "+in.View())
		keys = "Enter saves · Esc cancels"
	case editMulti:
		lines = append(lines, "")
		for i, c := range e.choices {
			box := "[ ]"
			if e.marked[i] {
				box = "[x]"
			}
			row := " " + map[bool]string{true: th.Icons.Pointer, false: " "}[i == e.at] + " " + box + " " + c
			if i == e.at {
				row = th.Selected(pane.Fit(pane.Strip(row), w))
			}
			lines = append(lines, row)
		}
		keys = "Space ticks · Enter saves · Esc cancels"
	case editReview:
		var text []string
		for _, l := range strings.Split(strings.TrimSpace(e.text), "\n") {
			text = append(text, pane.Wrap(" "+l, max(w, 4), 3)...)
		}
		room := max(h-len(lines)-3, 1)
		top := clamp(e.top, 0, len(text)-room)
		lines = append(lines, text[top:min(top+room, len(text))]...)
		keys = "y trust · p trust and pin the sha256 for CI · Esc cancels"
		if e.trusted {
			keys = "y trusts it again · p also pins the sha256 for CI · f forgets it · Esc closes"
		}
		if len(text) > room {
			keys += " · Up/Down read on"
		}
	}
	if e.err != "" {
		lines = append(lines, "")
		lines = append(lines, indent(pane.Wrap(th.Failure(e.err), max(w-2, 4), 0))...)
	}
	for len(lines) < h-1 {
		lines = append(lines, "")
	}
	lines = lines[:min(len(lines), max(h-1, 0))]
	return append(lines, " "+th.Muted(fmt.Sprint(keys)))
}

// shownValue is what a setting is now: the value written in smith.toml, or,
// when none is, the default in the muted colour, so a set value stands out
// from one that is only the default. A setting whose default is not known is
// shown as the word "default".
func shownValue(th theme.Theme, st Setting) string {
	switch {
	case st.Set:
		return st.Value
	case st.Default != "":
		return th.Muted(st.Default)
	}
	return th.Muted("default")
}
