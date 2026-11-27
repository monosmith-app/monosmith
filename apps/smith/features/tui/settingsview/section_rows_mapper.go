package settingsview

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/choicerow"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pathfinder"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
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
		return m.settingRow(th, m.settingAt(it), w)
	case kindFact:
		f := m.proj.Facts[it.idx]
		pw := m.pathWidth(w)
		cols := m.columns(w - 3 - pw)
		row := "   " + pane.Cell(f.Key, pw) + pane.Cell(f.Value, cols[0])
		if len(cols) > 1 {
			row += pane.Cell(th.Muted(f.Source), cols[1])
		}
		if len(cols) > 2 {
			row += th.Muted(words.Say("settingspage.type.fact"))
		}
		return row
	case kindRecipe:
		r := m.snap.Recipes[it.idx]
		if r.Error != "" {
			return "   " + th.Failure(th.Icons.Failed) + " " + pane.Cell(r.Path, 28) + th.Failure(words.Say("settingsview.row.recipe.invalid"))
		}
		return "   " + th.Success(th.Icons.Passed) + " " + pane.Cell(r.Path, 28) + th.Muted(strings.TrimSpace(r.Name+" "+r.Version))
	case kindForm:
		f := m.forms[it.idx]
		return "   " + pane.Cell(f.Title, 28) + th.Muted(f.Hint)
	case kindWorkspaceRecipe, kindRecipeWizard:
		return m.recipeActionRow(th, it)
	case kindUpdate:
		return "   " + pane.Cell(words.Say("settingsview.update.row", "version", m.upd.Current), 22) + m.updateState(th)
	case kindUpdateMute:
		return "   " + choicerow.Box(m.upd.Muted, false) + " " + words.Say("settingsview.update.mute")
	}
	return ""
}

// settingRow is a setting as a row of the table: its name, what it is now (a
// switch as ● on or ○ off, one of a fixed set with ▾), its default and its type,
// as far as the width allows.
func (m Model) settingRow(th theme.Theme, st Setting, w int) string {
	pw := m.pathWidth(w)
	cols := m.columns(w - 3 - pw)
	row := "   " + pane.Cell(st.label(), pw) + pane.Cell(tableValue(th, st), cols[0])
	if len(cols) > 1 {
		def := st.Default
		if def == "" {
			def = words.Say("settingspage.value.undecided")
		}
		row += pane.Cell(th.Muted(def), cols[1])
	}
	if len(cols) > 2 {
		row += th.Muted(typeName(st))
	}
	return row
}

// tableValue is the Value cell of a setting: a switch shows its state with the
// mark of a radio, a value of a fixed set says it is picked from a list.
func tableValue(th theme.Theme, st Setting) string {
	switch {
	case st.Type == "boolean" && !st.Set && st.Default == "":
		return th.Muted(th.Icons.Queued + " " + words.Say("settingspage.value.undecided"))
	case st.Type == "boolean" && boolValue(st):
		v := th.Icons.Dot + " " + words.Say("settingsview.row.on")
		if !st.Set {
			return th.Muted(v)
		}
		return th.Success(v)
	case st.Type == "boolean":
		return th.Muted(th.Icons.Queued + " " + words.Say("settingsview.row.off"))
	case st.Type == "string" && len(st.Choices) > 0:
		return shownValue(th, st) + " " + th.Muted(th.Icons.Open)
	}
	return shownValue(th, st)
}

// typeName is the Type cell of a setting: switch, choice, list, number or text.
func typeName(st Setting) string {
	switch {
	case st.Type == "boolean":
		return words.Say("settingspage.type.switch")
	case len(st.Choices) > 0 && st.Type == "array":
		return words.Say("settingspage.type.ticks")
	case len(st.Choices) > 0:
		return words.Say("settingspage.type.choice")
	case st.Type == "array":
		return words.Say("settingspage.type.list")
	case st.Type == "integer" || st.Type == "number":
		return words.Say("settingspage.type.number")
	}
	return words.Say("settingspage.type.text")
}

// configLine is the command line that writes a setting with values, or that
// resets it when values is empty: what Save runs, in this process.
func configLine(path string, values []string) string {
	if len(values) == 0 || (len(values) == 1 && values[0] == "") {
		return "$ smith config unset " + path // words:literal
	}
	return "$ smith config set " + path + " " + strings.Join(values, " ") // words:literal
}

// pathWidth is the width of the setting names: the longest, up to half the pane.
func (m Model) pathWidth(w int) int {
	n := 0
	if m.section == Project {
		for _, f := range m.proj.Facts {
			n = max(n, pane.Width(f.Key))
		}
		for _, s := range m.proj.Settings {
			n = max(n, pane.Width(s.label()))
		}
		return min(n+2, max(w/2, 12))
	}
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
		return oneLine(m.settingAt(it).Explanation)
	case kindFact:
		return oneLine(m.proj.Facts[it.idx].Explanation)
	case kindTheme, kindIcons:
		return words.Say("settingsview.summary.style")
	case kindUpdate, kindUpdateMute:
		return words.Say("settingsview.summary.update")
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
	case kindUpdate:
		add(th.Bold(words.Say("settingsview.update.title")), words.Say("settingsview.update.about"))
		switch {
		case m.upd.Available:
			add(th.Failure(words.Say("settingsview.update.available", "latest", m.upd.Latest, "current", m.upd.Current)))
			if m.upd.Runnable {
				add(th.Muted(words.Say("settingsview.update.runs"))+m.upd.Command, th.Muted(words.Say("settingsview.update.enter_update")))
			} else {
				add(words.Say("settingsview.update.unknown", "latest", m.upd.Latest))
			}
		case m.upd.Latest != "":
			add(th.Success(words.Say("settingsview.update.current", "version", m.upd.Current)))
			add(th.Muted(words.Say("settingsview.update.enter_check")))
		default:
			add(th.Muted(words.Say("settingsview.update.notchecked")), th.Muted(words.Say("settingsview.update.enter_check")))
		}
	case kindUpdateMute:
		add(th.Bold(words.Say("settingsview.update.mute")), words.Say("settingsview.update.mute_about"), th.Muted(words.Say("settingsview.update.mute_enter")))
	case kindTheme:
		add(m.previewLines(th, it, w)...)
	case kindIcons:
		add(m.previewLines(th, it, w)...)
	case kindFact:
		f := m.proj.Facts[it.idx]
		add(th.Bold(f.Key)+th.Muted("  "+words.Say("settingsview.fact.tag")), f.Explanation)
		add(th.Muted(words.Say("settingsview.setting.now")) + f.Value)
		add(th.Muted(words.Say("settingsview.fact.source", "file", f.Source)))
	case kindSetting:
		st := m.settingAt(it)
		// The box this sits in is titled by the path, and the table says the type.
		add(st.Explanation)
		if st.Default != "" {
			add(th.Muted(words.Say("settingsview.setting.default")) + st.Default)
		}
		if st.Reason != "" {
			add(th.Muted(words.Say("settingsview.setting.why")) + st.Reason)
		}
		if len(st.Choices) > 0 {
			add(th.Muted(words.Say("settingsview.setting.choices")))
			for _, o := range st.Choices {
				add("  " + optionText(th, o))
			}
		}
		now := shownValue(th, st)
		if !st.Set {
			now += th.Muted(words.Say("settingsview.setting.isdefault"))
		}
		keys := enterKey(st) + resetTail(st)
		add(th.Muted(words.Say("settingsview.setting.now"))+now, th.Muted(keys))
		add("", th.Faint(configLine(st.Path, nowValues(st))))
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
	case kindWorkspaceRecipe, kindRecipeWizard:
		add(m.recipeActionDetail(th, it)...)
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
	if e.about != "" {
		add(th.Muted(e.about))
	}
	var keys string
	switch e.kind {
	case editRecipeName:
		lines = append(lines, m.recipeNameLines(th, w)...)
		keys = words.Say("settingsview.keys.install")
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
		for i, o := range e.setting.Choices {
			r := choicerow.Row{Lead: choicerow.Radio(th, i == e.current) + " ", Selected: i == e.at}
			r.Label, r.Hint = optionRow(o)
			// Every line of a wrapped row answers a click on that option.
			for _, l := range choicerow.Lines(th, r, max(w-1, 1)) {
				rowOf[len(lines)] = i
				lines = append(lines, " "+l)
			}
		}
		keys = words.Say("settingsview.keys.choose")
	case editMulti:
		lines = append(lines, "")
		for i, o := range e.setting.Choices {
			r := choicerow.Ticks("", e.marked[i], false)
			r.Label, r.Hint = optionRow(o)
			r.Selected = i == e.at
			// Each tick box row answers a click (see Click), as in every list of boxes.
			for _, l := range choicerow.Lines(th, r, max(w-1, 1)) {
				rowOf[len(lines)] = i
				lines = append(lines, " "+l)
			}
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
	if e.kind == editText || e.kind == editChoose || e.kind == editMulti {
		// The command line the change stands for, as the person would type it.
		lines = append(lines, "", " "+th.Faint(configLine(e.setting.Path, e.values())))
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

// optionRow is an option as a row of a pick list: the word, marked when it is
// the default, and what it does as the muted hint.
func optionRow(o Option) (label, hint string) {
	label = o.Value
	if o.Default {
		label += words.Say("settingsview.setting.isdefault")
	}
	return label, o.Description
}

// optionText is an option on one line of the detail pane.
func optionText(th theme.Theme, o Option) string {
	label, hint := optionRow(o)
	if hint == "" {
		return label
	}
	return label + th.Muted(" - "+hint)
}

// enterKey says what Enter does to a setting: a switch or a value of a fixed
// set is picked from a list, anything else is changed.
func enterKey(st Setting) string {
	switch {
	case st.Path == "cache.remote":
		return words.Say("settingsview.setting.keys.cache-setup")
	case st.Type == "boolean" || (st.Type == "string" && len(st.Choices) > 0):
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

// nowValues are the values a setting has now, for its command line: the ones
// written, else its default.
func nowValues(st Setting) []string {
	if st.Set && len(st.Current) > 0 {
		return st.Current
	}
	if st.Type == "boolean" && st.Default != "" {
		return []string{st.Default}
	}
	if st.Set {
		return []string{st.Value}
	}
	return []string{"<" + typeName(st) + ">"}
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

// updateState is the state of the update row: a newer smith in red, the newest
// one that is installed in green, or that nothing was checked yet.
func (m Model) updateState(th theme.Theme) string {
	switch {
	case m.upd.Available:
		return th.Failure(words.Say("settingsview.update.state_available", "latest", m.upd.Latest))
	case m.upd.Latest != "":
		return th.Success(words.Say("settingsview.update.state_current"))
	}
	return th.Muted(words.Say("settingsview.update.state_unknown"))
}
