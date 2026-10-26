package settingsview

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

var errUnavailable = errors.New("this smith cannot do that here")

// Key gives the tab a key press.
func (m Model) Key(msg tea.KeyMsg) (Model, Result) {
	if m.ed != nil {
		return m.editorKey(msg)
	}
	name := msg.String()
	switch name {
	case "up", "down", "pgup", "pgdown", "home", "end":
		return m.move(name), Result{}
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

// move moves the cursor of the rows, or the section while the keyboard is on the list of sections.
func (m Model) move(key string) Model {
	m.note = ""
	if !m.onItems {
		at := int(m.section)
		switch key {
		case "up":
			at--
		case "down":
			at++
		case "home":
			at = 0
		case "end":
			at = len(Sections) - 1
		}
		m.section = Section(max(min(at, len(Sections)-1), 0))
		return m
	}
	n := len(m.items(m.section))
	c := m.cursor[m.section]
	switch key {
	case "up":
		c--
	case "down":
		c++
	case "pgup":
		c -= 8
	case "pgdown":
		c += 8
	case "home":
		c = 0
	case "end":
		c = n - 1
	}
	m.cursor[m.section] = max(min(c, n-1), 0)
	return m
}

// activate does what Enter or Space does on the row under the cursor.
func (m Model) activate() (Model, Result) {
	it, ok := m.current()
	if !ok {
		return m, Result{}
	}
	m.note = ""
	switch it.kind {
	case kindPlugin:
		p := m.snap.Plugins[it.idx]
		return m, Result{Cmd: m.setPlugin(p.Name, !p.Enabled)}
	case kindExternal:
		return m.askTrust()
	case kindInstall:
		if m.b.Install == nil {
			m.note = "Not in this smith yet: run  smith plugin install <name|path>  in a terminal when it is."
			return m, Result{}
		}
		m.ed = newInstall()
	case kindTheme:
		return m, Result{Cmd: m.saveStyle(normalise(it.key), m.savedIcons)}
	case kindIcons:
		return m, Result{Cmd: m.saveStyle(m.savedTheme, normalise(it.key))}
	case kindSetting:
		return m.editSetting(m.snap.Settings[it.idx])
	case kindForm:
		return m, Result{Form: it.key}
	case kindInstalled:
		m.note = "Installed here; a workspace uses it when smith.toml declares it under [plugins.external]."
	case kindRecipe:
		m.note = "Recipes are chosen in smith.toml ([workspace] recipe); smith recipe check <file> checks one."
	}
	return m, Result{}
}

// editSetting changes a setting the way its type asks: a switch flips, a
// fixed list is ticked, anything else is typed.
func (m Model) editSetting(st Setting) (Model, Result) {
	switch {
	case st.Type == "boolean":
		return m, Result{Cmd: m.setSetting(st.Path, []string{fmt.Sprint(!boolValue(st))})}
	case st.Type == "array" && len(st.Choices) > 0:
		m.ed = newPicker(st)
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
	return m, Result{Cmd: func() tea.Msg {
		r, err := review(name)
		return Reviewed{Name: name, Review: r, Err: err}
	}}
}

func (m Model) forgetTrust() (Model, Result) {
	it, ok := m.current()
	if !ok || it.kind != kindExternal || !m.onItems || !m.snap.External[it.idx].Trusted {
		return m, Result{}
	}
	return m, Result{Cmd: m.untrust(m.snap.External[it.idx].Name)}
}

func (m Model) resetSetting() (Model, Result) {
	it, ok := m.current()
	if !ok || it.kind != kindSetting || !m.onItems {
		return m, Result{}
	}
	st := m.snap.Settings[it.idx]
	if !st.Set {
		m.note = st.Path + " already has its default"
		return m, Result{}
	}
	return m, Result{Cmd: m.unsetSetting(st.Path)}
}

func unavailable() tea.Cmd { return func() tea.Msg { return Applied{Err: errUnavailable} } }

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

func (m Model) setPlugin(name string, enable bool) tea.Cmd {
	set := m.b.SetPlugin
	if set == nil {
		return unavailable()
	}
	return m.run(func() (string, error) {
		if err := set(name, enable); err != nil {
			return "", err
		}
		return name + map[bool]string{true: " is on", false: " is off"}[enable], nil
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

func (m Model) install(source string) tea.Cmd {
	if m.b.Install == nil {
		return unavailable()
	}
	return m.run(func() (string, error) { return m.b.Install(source) })
}

// saveStyle keeps the choice and reports the theme to draw with: the saved
// choice under the environment variables that win over it.
func (m Model) saveStyle(themeName, icons string) tea.Cmd {
	st := m.b.Style
	return func() tea.Msg {
		if st.Save == nil {
			return StyleSaved{Err: errUnavailable}
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
