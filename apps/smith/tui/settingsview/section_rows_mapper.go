package settingsview

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pathfinder"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// themeAbout and iconsAbout are the catalog key of what each choice of the Style
// section means.
var (
	themeAbout = map[string]string{
		theme.Auto: "settingsview.theme.auto",
		"dark":     "settingsview.theme.dark",
		"light":    "settingsview.theme.light",
		"plain":    "settingsview.theme.plain",
	}
	iconsAbout = map[string]string{
		theme.Auto: "settingsview.icons.auto",
		"unicode":  "settingsview.icons.unicode",
		"ascii":    "settingsview.icons.ascii",
	}
)

// itemRow is one row of a list.
func (m Model) itemRow(th theme.Theme, it item, w int) string {
	switch it.kind {
	case kindPlugin:
		p := m.snap.Plugins[it.idx]
		state := th.Muted(words.Say("settingsview.row.off"))
		if p.Enabled {
			state = th.Success(words.Say("settingsview.row.on"))
		}
		return "   " + choicerow.Box(p.Enabled, false) + " " + pane.Cell(p.Name, 14) + state
	case kindExternal:
		p := m.snap.External[it.idx]
		mark, state := th.Success(th.Icons.Passed), words.Say("settingsview.row.trusted")
		switch {
		case p.Problem != "":
			mark, state = th.Failure(th.Icons.Failed), words.Say("settingsview.row.cannotstart", "problem", p.Problem)
		case !p.Trusted:
			mark, state = th.Warning(th.Icons.Warning), words.Say("settingsview.row.untrusted")
		case p.Pinned:
			state = words.Say("settingsview.row.pinned")
		}
		return "   " + mark + " " + pane.Cell(p.Name, 14) + pane.Cell(state, 24) + th.Muted(p.Source)
	case kindInstalled:
		p := m.snap.Installed[it.idx]
		if p.Problem != "" {
			return "   " + th.Failure(th.Icons.Failed) + " " + pane.Cell(p.Name, 14) + th.Failure(words.Say("settingsview.row.invalid", "problem", p.Problem))
		}
		return "     " + pane.Cell(p.Name, 14) + pane.Cell(p.Version, 10) + th.Muted(p.Dir)
	case kindInstall:
		return "   " + th.Accent(words.Say("settingsview.row.install")) + "  " + th.Muted("smith plugin install <source>") // words:literal
	case kindSearch:
		return "   " + th.Accent(words.Say("settingsview.row.search")) + "  " + th.Muted("smith plugin search") // words:literal
	case kindArchitecture:
		a := m.snap.Architectures[it.idx]
		return "   " + choicerow.Radio(th, a.Name == m.snap.Architecture) + " " + pane.Cell(a.Name, 16) + th.Muted(a.Meaning)
	case kindTheme:
		return "   " + choicerow.Radio(th, normalise(it.key) == m.savedTheme) + " " + pane.Cell(it.key, 8) + th.Muted(words.Say(themeAbout[it.key]))
	case kindIcons:
		return "   " + choicerow.Radio(th, normalise(it.key) == m.savedIcons) + " " + pane.Cell(it.key, 8) + th.Muted(words.Say(iconsAbout[it.key]))
	case kindSetting:
		st := m.snap.Settings[it.idx]
		if st.Type == "boolean" {
			// A switch shows its state as a box, then the word the file holds.
			return "   " + pane.Cell(st.Path, m.pathWidth(w)) + choicerow.Box(boolValue(st), false) + " " + shownValue(th, st)
		}
		return "   " + pane.Cell(st.Path, m.pathWidth(w)) + shownValue(th, st)
	case kindRecipe:
		r := m.snap.Recipes[it.idx]
		if r.Error != "" {
			return "   " + th.Failure(th.Icons.Failed) + " " + pane.Cell(r.Path, 28) + th.Failure(words.Say("settingsview.row.recipe.invalid"))
		}
		return "   " + th.Success(th.Icons.Passed) + " " + pane.Cell(r.Path, 28) + th.Muted(strings.TrimSpace(r.Name+" "+r.Version))
	case kindForm:
		f := m.forms[it.idx]
		return "   " + pane.Cell(f.Title, 28) + th.Muted(f.Hint)
	}
	return ""
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
		return words.Say("settingsview.summary.plugin", "name", it.key)
	case kindExternal:
		return words.Say("settingsview.summary.external")
	case kindSetting:
		return oneLine(m.snap.Settings[it.idx].Explanation)
	case kindTheme, kindIcons:
		return words.Say("settingsview.summary.style")
	case kindArchitecture:
		return words.Say("settingsview.summary.architecture")
	}
	return ""
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// emptyList is what a section with no rows says.
func (m Model) emptyList(th theme.Theme, w int) []string {
	if m.section == Recipes {
		return append([]string{""}, indent(pane.Wrap(th.Muted(words.Say("settingsview.empty.recipes")), w-2, 0))...)
	}
	return []string{"", " " + th.Muted(words.Say("settingsview.empty.list"))}
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
		add(th.Bold(it.key)+th.Muted(words.Say("settingsview.plugin.support")),
			words.Say("settingsview.plugin.about"),
			th.Muted(words.Say("settingsview.plugin.addons", "tools", strings.Join(m.snap.Addons, ", "), "name", it.key)))
	case kindExternal:
		p := m.snap.External[it.idx]
		add(th.Bold(p.Name)+th.Muted(words.Say("settingsview.external.from", "source", p.Source)), th.Muted(words.Say("settingsview.external.runs"))+p.Command)
		sees := words.Say("settingsview.external.sees.baseline")
		if len(p.Env) > 0 {
			sees = words.Say("settingsview.external.sees.plus", "env", strings.Join(p.Env, ", "))
		}
		add(th.Muted(words.Say("settingsview.external.sees")) + sees)
		switch {
		case p.Problem != "":
			add(th.Failure(p.Problem))
		case p.Trusted:
			add(words.Say("settingsview.external.trusted", "state", th.Success(words.Say("settingsview.external.trusted.word"))))
		default:
			add(words.Say("settingsview.external.untrusted", "state", th.Warning(words.Say("settingsview.external.untrusted.word")), "detail", p.Detail))
		}
	case kindInstalled:
		p := m.snap.Installed[it.idx]
		add(th.Bold(p.Name)+th.Muted("  "+p.Version), th.Muted(p.Dir))
		if p.Problem != "" {
			add(th.Failure(p.Problem))
		} else {
			add(words.Say("settingsview.installed.about", "name", p.Name))
		}
		add(th.Muted(words.Say("settingsview.installed.path", "path", strings.Join(m.snap.SearchPath, ", "))))
	case kindInstall:
		add(th.Bold(words.Say("settingsview.install.title")), words.Say("settingsview.install.from"),
			words.Say("settingsview.install.detail"))
	case kindSearch:
		add(th.Bold(words.Say("settingsview.search.title")), words.Say("settingsview.search.from"),
			words.Say("settingsview.search.detail"))
	case kindArchitecture:
		a := m.snap.Architectures[it.idx]
		add(th.Bold(a.Name)+th.Muted("  "+a.Meaning),
			words.Say("settingsview.architecture.about"),
			words.Say("settingsview.architecture.enter", "writes", a.Writes))
		if a.AppliesTo != "" {
			add(th.Bold(words.Say("settingsview.architecture.appliesto")) + a.AppliesTo)
		}
		if a.Here != "" {
			add(th.Bold(words.Say("settingsview.architecture.here")) + a.Here)
		}
		if a.Warning != "" {
			add(th.Failure(a.Warning))
		}
		add(th.Muted(words.Say("settingsview.architecture.same", "name", a.Name)))
	case kindTheme:
		add(m.previewLines(th, it, w)...)
	case kindIcons:
		add(m.previewLines(th, it, w)...)
	case kindSetting:
		st := m.snap.Settings[it.idx]
		add(th.Bold(st.Path)+th.Muted("  "+st.Type), st.Explanation)
		if st.Default != "" {
			add(th.Muted(words.Say("settingsview.setting.default")) + st.Default)
		}
		if st.Reason != "" {
			add(th.Muted(words.Say("settingsview.setting.why")) + st.Reason)
		}
		if len(st.Choices) > 0 {
			add(th.Muted(words.Say("settingsview.setting.choices")) + strings.Join(st.Choices, ", "))
		}
		now := shownValue(th, st)
		if !st.Set {
			now += th.Muted(words.Say("settingsview.setting.isdefault"))
		}
		keys := enterKey(st) + resetTail(st)
		add(th.Muted(words.Say("settingsview.setting.now"))+now, th.Muted(words.Say("settingsview.setting.command", "keys", keys, "path", st.Path)))
	case kindRecipe:
		r := m.snap.Recipes[it.idx]
		add(th.Bold(r.Path))
		if r.Error != "" {
			add(th.Failure(r.Error))
		} else {
			add(strings.TrimSpace(r.Name + " " + r.Version))
			if r.Extends != "" {
				add(th.Muted(words.Say("settingsview.recipe.extends")) + r.Extends)
			}
		}
		add(th.Muted(words.Say("settingsview.recipe.check", "path", r.Path)))
	case kindForm:
		f := m.forms[it.idx]
		add(th.Bold(f.Title), f.Hint, th.Muted(words.Say("settingsview.form.about")))
	}
	return out
}

// editorLines draws the open editor in w by h cells. rowOf maps each line of
// the path finder's rows to its row, for a click.
func (m Model) editorLines(th theme.Theme, w, h int) (lines []string, rowOf map[int]int) {
	e := m.ed
	rowOf = map[int]int{}
	add := func(s string) { lines = append(lines, pane.Wrap(" "+s, max(w, 4), 1)...) }
	lines = append(lines, "")
	add(th.Bold(e.title))
	if e.about != "" {
		add(th.Muted(e.about))
	}
	var keys string
	switch e.kind {
	case editText, editInstall, editSearch:
		if e.setting.Default != "" {
			add(th.Muted(words.Say("settingsview.setting.default.line", "default", e.setting.Default)))
		}
		if e.setting.Reason != "" {
			add(th.Muted(words.Say("settingsview.setting.why.line", "reason", e.setting.Reason)))
		}
		if e.setting.Type == "array" {
			add(th.Muted(words.Say("settingsview.editor.commas")))
		}
		in := e.input
		in.Width = max(w-6, 4)
		lines = append(lines, "", " "+in.View())
		keys = words.Say("settingsview.keys.save")
		if e.finder != nil {
			for _, l := range e.finder.Rows(th, max(w-1, 4), max(min(pathfinder.DefaultRows, h-len(lines)-5), 2)) {
				if l.Index >= 0 {
					rowOf[len(lines)] = l.Index
				}
				lines = append(lines, " "+l.Text)
			}
			keys = words.Say("settingsview.keys.finder", "keys", keys)
		}
		if e.kind == editInstall {
			argv, _ := installArgv(e.input.Value())
			lines = append(lines, "", " "+th.Muted("smith "+strings.Join(argv, " "))) // words:literal
			keys = words.Say("settingsview.keys.install")
		}
		if e.kind == editSearch {
			keys = words.Say("settingsview.keys.search")
		}
	case editChoose:
		lines = append(lines, "")
		for i, c := range e.choices {
			row := " " + choicerow.Cursor(th, i == e.at) + choicerow.Radio(th, i == e.current) + " " + c
			if i == e.at {
				row = th.Selected(pane.Fit(pane.Strip(row), w))
			}
			rowOf[len(lines)] = i
			lines = append(lines, row)
		}
		keys = words.Say("settingsview.keys.choose")
	case editMulti:
		lines = append(lines, "")
		for i, c := range e.choices {
			row := " " + choicerow.Render(th, c, e.marked[i], i == e.at, false)
			if i == e.at {
				row = th.Selected(pane.Fit(pane.Strip(row), w))
			}
			lines = append(lines, row)
		}
		keys = words.Say("settingsview.keys.multi")
	case editReview:
		var text []string
		for _, l := range strings.Split(strings.TrimSpace(e.text), "\n") {
			text = append(text, pane.Wrap(" "+l, max(w, 4), 3)...)
		}
		room := max(h-len(lines)-3, 1)
		top := pane.Clamp(e.top, 0, len(text)-room)
		lines = append(lines, text[top:min(top+room, len(text))]...)
		keys = words.Say("settingsview.keys.review")
		if e.trusted {
			keys = words.Say("settingsview.keys.review.again")
		}
		if len(text) > room {
			keys = words.Say("settingsview.keys.more", "keys", keys)
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
	return append(lines, " "+th.Muted(fmt.Sprint(keys))), rowOf
}

// enterKey says what Enter does to a setting: a switch flips, a value of a fixed
// set is picked, anything else is changed.
func enterKey(st Setting) string {
	switch {
	case st.Path == "cache.remote":
		return words.Say("settingsview.setting.keys.cache-setup")
	case st.Type == "boolean":
		return words.Say("settingsview.setting.keys.flip")
	case st.Type == "string" && len(st.Choices) > 0:
		return words.Say("settingsview.setting.keys.pick")
	}
	return words.Say("settingsview.setting.keys.change")
}

// resetTail is the sentence about the u key after the one about Enter: it is
// there only while a value is written, as the key does nothing otherwise.
func resetTail(st Setting) string {
	if !st.Set {
		return ""
	}
	return words.Say("settingsview.setting.keys.reset")
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
	return th.Muted(words.Say("settingsview.value.default"))
}
