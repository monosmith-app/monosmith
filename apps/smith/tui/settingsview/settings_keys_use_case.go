package settingsview

import (
	"errors"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// errUnavailable is the error of a change this smith has no backend for.
func errUnavailable() error { return errors.New(words.Say("settingsview.error.unavailable")) }

// Key gives the tab a key press.
func (m Model) Key(msg tea.KeyMsg) (Model, Result) {
	if m.busy != "" {
		if res, handled := m.busyKey(msg); handled {
			return res.model, res.Result
		}
	}
	if m.ed != nil {
		return m.editorKey(msg)
	}
	name := msg.String()
	if nav := commands.NavNamed(name); nav.Moves() {
		return m.move(nav), Result{}
	}
	switch name {
	case "tab":
		m.onItems = !m.onItems
	case "shift+tab", "left", "backspace":
		m.onItems = false
	case "right":
		m.onItems = true
	case "esc":
		m.onItems = false
	case "enter", " ":
		if !m.onItems {
			m.onItems = true
			return m, Result{}
		}
		return m.activate()
	case "t":
		return m.askTrust()
	case "f":
		return m.forgetTrust()
	case "u":
		return m.resetSetting()
	}
	return m, Result{}
}

// busyAnswer is the tab and the result of a key pressed while work runs.
type busyAnswer struct {
	model Model
	Result
}

// busyKey handles a key while a change runs: moving still works, Esc stops a
// review, and any other key says why it waits. handled is false for a key the
// tab treats as usual.
func (m Model) busyKey(msg tea.KeyMsg) (busyAnswer, bool) {
	name := msg.String()
	nav := commands.NavNamed(name)
	switch {
	case name == "esc":
		if m.stoppable {
			m.busy, m.ticket = "", m.ticket+1
			return busyAnswer{m, Result{Note: words.Say("settingsview.busy.cancelled")}}, true
		}
		return busyAnswer{m, Result{Note: words.Say("settingsview.busy.unstoppable")}}, true
	case nav.Moves() || nav == commands.NavNext || nav == commands.NavPrev || name == "left" || name == "right":
		if m.ed == nil {
			return busyAnswer{}, false
		}
	}
	return busyAnswer{m, Result{Note: words.Say("settingsview.busy.wait", "task", m.busy)}}, true
}

// move moves the cursor of the rows, or the section while the keyboard is on the list of sections.
func (m Model) move(key commands.NavAction) Model {
	m.note, m.confirm = "", ""
	if !m.onItems {
		at := int(m.section)
		switch key {
		case commands.NavUp:
			at--
		case commands.NavDown:
			at++
		case commands.NavHome:
			at = 0
		case commands.NavEnd:
			at = len(Sections) - 1
		}
		m.section = Section(max(min(at, len(Sections)-1), 0))
		return m
	}
	n := len(m.items(m.section))
	c := m.cursor[m.section]
	switch key {
	case commands.NavUp:
		c--
	case commands.NavDown:
		c++
	case commands.NavPageUp:
		c -= 8
	case commands.NavPageDown:
		c += 8
	case commands.NavHome:
		c = 0
	case commands.NavEnd:
		c = n - 1
	}
	m.cursor[m.section] = max(min(c, n-1), 0)
	return m
}

// activate does what Enter or Space does on the row under the cursor.
func (m Model) activate() (Model, Result) {
	it, ok := m.current()
	if !ok || m.busy != "" {
		return m, Result{}
	}
	m.note = ""
	switch it.kind {
	case kindPlugin:
		p := m.snap.Plugins[it.idx]
		return m.switchPlugin(p.Name, !p.Enabled)
	case kindExternal:
		return m.askTrust()
	case kindInstall:
		m.ed = newInstall()
	case kindSearch:
		m.ed = newSearch()
	case kindArchitecture:
		a := m.snap.Architectures[it.idx]
		if a.Warning != "" && m.confirm != a.Name {
			// It would check nothing here: say so, and apply on the second Enter.
			m.confirm = a.Name
			m.note = words.Say("settingsview.architecture.confirm", "name", a.Name)
			return m, Result{}
		}
		m.confirm = ""
		m = m.begin(words.Say("settingsview.busy.architecture", "name", it.key), false)
		return m, Result{Cmd: m.setArchitecture(it.key)}
	case kindTheme:
		m = m.begin(words.Say("settingsview.busy.style"), false)
		return m, Result{Cmd: m.saveStyle(normalise(it.key), m.savedIcons)}
	case kindIcons:
		m = m.begin(words.Say("settingsview.busy.style"), false)
		return m, Result{Cmd: m.saveStyle(m.savedTheme, normalise(it.key))}
	case kindSetting:
		if m.snap.Settings[it.idx].Path == "cache.remote" {
			// The remote is chosen with the setup, which finds it, shows each
			// command first, and writes the setting only after a yes.
			return m, Result{CacheSetup: true}
		}
		return m.editSetting(m.snap.Settings[it.idx])
	case kindForm:
		return m, Result{Form: it.key}
	case kindUpdate:
		return m.activateUpdate()
	case kindUpdateMute:
		return m.toggleMuted()
	case kindInstalled:
		m.note = words.Say("settingsview.note.installed")
	case kindRecipe:
		m.note = words.Say("settingsview.note.recipe")
	}
	return m, Result{}
}

// editSetting changes a setting the way its type asks: a switch flips, a
// fixed list is ticked, one value of a fixed set is picked, anything else is typed.
func (m Model) editSetting(st Setting) (Model, Result) {
	switch {
	case st.Type == "boolean":
		m = m.begin(words.Say("settingsview.busy.saving", "path", st.Path), false)
		return m, Result{Cmd: m.setSetting(st.Path, []string{fmt.Sprint(!boolValue(st))})}
	case st.Type == "array" && len(st.Choices) > 0:
		m.ed = newPicker(st)
	case st.Type == "string" && len(st.Choices) > 0:
		m.ed = newChooser(st)
	default:
		m.ed = newTextEditor(st)
	}
	return m, Result{}
}

// boolValue is a switch as it is now: written, or its default.
func boolValue(st Setting) bool {
	if st.Set {
		return st.Value == "true"
	}
	return st.Default == "true"
}

func (m Model) askTrust() (Model, Result) {
	it, ok := m.current()
	if !ok || it.kind != kindExternal || !m.onItems {
		return m, Result{}
	}
	review := m.b.Review
	if review == nil {
		return m, Result{Cmd: unavailable()}
	}
	name := m.snap.External[it.idx].Name
	m = m.begin(words.Say("settingsview.busy.reading", "name", name), true)
	ticket := m.ticket
	return m, Result{Cmd: func() tea.Msg {
		r, err := review(name)
		return Reviewed{Ticket: ticket, Name: name, Review: r, Err: err}
	}}
}

func (m Model) forgetTrust() (Model, Result) {
	it, ok := m.current()
	if !ok || it.kind != kindExternal || !m.onItems || !m.snap.External[it.idx].Trusted {
		return m, Result{}
	}
	m = m.begin(words.Say("settingsview.busy.forgetting", "name", m.snap.External[it.idx].Name), false)
	return m, Result{Cmd: m.untrust(m.snap.External[it.idx].Name)}
}

func (m Model) resetSetting() (Model, Result) {
	it, ok := m.current()
	if !ok || it.kind != kindSetting || !m.onItems {
		return m, Result{}
	}
	st := m.snap.Settings[it.idx]
	if !st.Set {
		m.note = words.Say("settingsview.note.default", "path", st.Path)
		return m, Result{}
	}
	m = m.begin(words.Say("settingsview.busy.resetting", "path", st.Path), false)
	return m, Result{Cmd: m.unsetSetting(st.Path)}
}

func unavailable() tea.Cmd { return func() tea.Msg { return Applied{Err: errUnavailable()} } }

// run does a change off the update loop and reads the workspace again after it.
func (m Model) run(op func() (string, error)) tea.Cmd {
	load := m.b.Load
	return func() tea.Msg {
		note, err := op()
		if err != nil {
			return Applied{Err: err}
		}
		a := Applied{Note: note}
		if load != nil {
			fresh := load()
			a.Fresh = &fresh
		}
		return a
	}
}

// switchPlugin turns a built-in plugin on or off in the background, and
// remembers the switch so the result can offer to undo it.
func (m Model) switchPlugin(name string, enable bool) (Model, Result) {
	m = m.begin(map[bool]string{true: words.Say("settingsview.busy.on", "name", name), false: words.Say("settingsview.busy.off", "name", name)}[enable], false)
	m.toggling = &PluginToggle{Name: name, Enabled: enable}
	return m, Result{Cmd: m.setPlugin(name, enable)}
}

// Revert switches a plugin back to the state before a switch the shell offered
// to undo.
//
// Parameters: t is the switch to undo: the plugin and the state it was switched to.
// Returns: the tab, busy with the switch back, and the work to run.
func (m Model) Revert(t PluginToggle) (Model, Result) {
	if m.busy != "" {
		return m, Result{}
	}
	return m.switchPlugin(t.Name, !t.Enabled)
}

func (m Model) setPlugin(name string, enable bool) tea.Cmd {
	set := m.b.SetPlugin
	if set == nil {
		return unavailable()
	}
	return m.run(func() (string, error) {
		warning, err := set(name, enable)
		if err != nil {
			return "", err
		}
		note := words.Say("settingsview.plugin.off", "name", name)
		if enable {
			note = words.Say("settingsview.plugin.on", "name", name)
		}
		if warning != "" {
			note = words.Say("settingsview.plugin.warning", "note", note, "warning", warning)
		}
		return note, nil
	})
}

func (m Model) setArchitecture(name string) tea.Cmd {
	set := m.b.SetArchitecture
	if set == nil {
		return unavailable()
	}
	return m.run(func() (string, error) {
		if err := set(name); err != nil {
			return "", err
		}
		return words.Say("settingsview.architecture.chosen", "name", name), nil
	})
}

func (m Model) setSetting(path string, values []string) tea.Cmd {
	if m.b.SetSetting == nil {
		return unavailable()
	}
	return m.run(func() (string, error) { return m.b.SetSetting(path, values) })
}

func (m Model) unsetSetting(path string) tea.Cmd {
	if m.b.UnsetSetting == nil {
		return unavailable()
	}
	return m.run(func() (string, error) { return m.b.UnsetSetting(path) })
}

func (m Model) trust(name string, pin bool) tea.Cmd {
	if m.b.Trust == nil {
		return unavailable()
	}
	return m.run(func() (string, error) { return m.b.Trust(name, pin) })
}

func (m Model) untrust(name string) tea.Cmd {
	if m.b.Untrust == nil {
		return unavailable()
	}
	return m.run(func() (string, error) { return m.b.Untrust(name) })
}

// saveStyle keeps the choice and reports the theme to draw with: the saved
// choice under the environment variables that win over it.
func (m Model) saveStyle(themeName, icons string) tea.Cmd {
	st := m.b.Style
	return func() tea.Msg {
		if st.Save == nil {
			return StyleSaved{Err: errUnavailable()}
		}
		if err := st.Save(themeName, icons); err != nil {
			return StyleSaved{Err: err}
		}
		getenv := st.Getenv
		if getenv == nil {
			getenv = func(string) string { return "" }
		}
		return StyleSaved{Theme: themeName, Icons: icons, Choice: theme.ChooseWith(getenv, st.Dark, themeName, icons)}
	}
}

// activateUpdate is Enter on the update row: with a newer smith known it asks the
// shell to offer the update; otherwise it asks the registry.
func (m Model) activateUpdate() (Model, Result) {
	if m.upd.Available {
		if !m.upd.Runnable {
			m.note = words.Say("settingsview.update.unknown", "latest", m.upd.Latest)
			return m, Result{}
		}
		return m, Result{Update: true}
	}
	check := m.b.Update.Check
	if check == nil {
		return m, Result{Note: words.Say("settingsview.update.unavailable"), Failed: true}
	}
	m = m.begin(words.Say("settingsview.busy.update"), false)
	return m, Result{Cmd: func() tea.Msg {
		info, err := check()
		return UpdateChecked{Info: info, Err: err}
	}}
}

// toggleMuted switches the reminder: the status bar and the first alert stay
// quiet, while the check, this section and smith update go on.
func (m Model) toggleMuted() (Model, Result) {
	set := m.b.Update.SetMuted
	if set == nil {
		return m, Result{Note: words.Say("settingsview.update.unavailable"), Failed: true}
	}
	muted := !m.upd.Muted
	if err := set(muted); err != nil {
		m.note, m.failed = words.Say("settingsview.update.failed", "reason", err.Error()), true
		return m, Result{Note: m.note, Failed: true}
	}
	m.upd.Muted = muted
	note := words.Say("settingsview.update.reminders_on")
	if muted {
		note = words.Say("settingsview.update.reminders_off")
	}
	return m, Result{Note: note, MutedChanged: true, Muted: muted}
}

// WithUpdate replaces what the Update section shows, when the shell learned it
// elsewhere (the first-run alert ticked the reminder box).
//
// Parameters: info is what smith knows now.
// Returns: the tab with it.
func (m Model) WithUpdate(info UpdateInfo) Model {
	m.upd = info
	return m
}
