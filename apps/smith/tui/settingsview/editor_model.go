package settingsview

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type editorKind int

const (
	// editText is one line of text: a string, a number, a list written with commas.
	editText editorKind = iota
	// editMulti picks any of a fixed set of values.
	editMulti
	// editReview is the review of an external plugin, with the keys that decide.
	editReview
	// editInstall asks what plugin to install.
	editInstall
)

// editor is what fills the right pane while a value is being changed.
type editor struct {
	kind  editorKind
	title string
	// about is what the thing does, shown above the field.
	about   string
	setting Setting
	input   textinput.Model
	// choices, marked and at are the picker's values, which are ticked and the cursor.
	choices []string
	marked  []bool
	at      int
	// text and top are the review and the first line shown of it.
	text    string
	top     int
	plugin  string
	trusted bool
	// err is why the last save was refused.
	err string
}

func newInput(value, placeholder string) textinput.Model {
	in := textinput.New()
	in.CharLimit = 200
	in.Placeholder = placeholder
	in.SetValue(value)
	in.CursorEnd()
	in.Focus()
	return in
}

// newTextEditor edits a string, number or list setting as one line; a list is
// written with commas.
func newTextEditor(st Setting) *editor {
	value := ""
	if st.Set {
		value = strings.Join(st.Current, ", ")
		if st.Type != "array" && len(st.Current) > 0 {
			value = st.Current[0]
		}
	}
	placeholder := st.Default
	if placeholder == "" {
		placeholder = "a " + st.Type
	}
	return &editor{kind: editText, title: st.Path, about: st.Explanation, setting: st, input: newInput(value, placeholder)}
}

// newPicker edits a list of fixed choices: Space ticks, Enter saves. When the
// setting is not set, what its default says is ticked.
func newPicker(st Setting) *editor {
	e := &editor{kind: editMulti, title: st.Path, about: st.Explanation, setting: st, choices: st.Choices, marked: make([]bool, len(st.Choices))}
	for i, c := range st.Choices {
		if st.Set {
			e.marked[i] = slices.Contains(st.Current, c)
		} else {
			e.marked[i] = strings.Contains(st.Default, `"`+c+`"`)
		}
	}
	return e
}

func newReview(name string, r TrustReview) *editor {
	return &editor{kind: editReview, title: "Trust " + name + "?", plugin: name, text: r.Text, trusted: r.Trusted}
}

func newInstall() *editor {
	return &editor{kind: editInstall, title: "Install a plugin", about: "A plugin's name, or the path to its directory. smith plugin install does the same.", input: newInput("", "name or path")}
}

// values are what the editor would save.
func (e editor) values() []string {
	switch e.kind {
	case editMulti:
		var out []string
		for i, c := range e.choices {
			if e.marked[i] {
				out = append(out, c)
			}
		}
		return out
	case editText:
		if e.setting.Type == "array" {
			var out []string
			for _, part := range strings.Split(e.input.Value(), ",") {
				if part = strings.TrimSpace(part); part != "" {
					out = append(out, part)
				}
			}
			return out
		}
	}
	return []string{strings.TrimSpace(e.input.Value())}
}

// editorKey gives the open editor a key.
func (m Model) editorKey(msg tea.KeyMsg) (Model, Result) {
	e := *m.ed
	name := msg.String()
	if name == "esc" {
		m.ed = nil
		return m, Result{}
	}
	switch e.kind {
	case editText, editInstall:
		if name == "enter" {
			e.err = ""
			m.ed = &e
			if e.kind == editInstall {
				return m, Result{Cmd: m.install(strings.TrimSpace(e.input.Value()))}
			}
			return m, Result{Cmd: m.setSetting(e.setting.Path, e.values())}
		}
		e.err = ""
		e.input, _ = e.input.Update(msg)
	case editMulti:
		switch name {
		case "up", "k":
			e.at = max(e.at-1, 0)
		case "down", "j":
			e.at = min(e.at+1, len(e.choices)-1)
		case " ":
			e.marked = slices.Clone(e.marked)
			e.marked[e.at] = !e.marked[e.at]
		case "enter":
			e.err = ""
			m.ed = &e
			return m, Result{Cmd: m.setSetting(e.setting.Path, e.values())}
		}
	case editReview:
		switch name {
		case "y":
			m.ed = &e
			return m, Result{Cmd: m.trust(e.plugin, false)}
		case "p":
			m.ed = &e
			return m, Result{Cmd: m.trust(e.plugin, true)}
		case "f", "u":
			if e.trusted {
				m.ed = &e
				return m, Result{Cmd: m.untrust(e.plugin)}
			}
		case "up", "k":
			e.top = max(e.top-1, 0)
		case "down", "j":
			e.top++
		case "pgup":
			e.top = max(e.top-8, 0)
		case "pgdown":
			e.top += 8
		}
	}
	m.ed = &e
	return m, Result{}
}

// keyOf is the key message of a key name the model turns a wheel notch into.
func keyOf(name string) tea.KeyMsg {
	return tea.KeyMsg{Type: map[string]tea.KeyType{"up": tea.KeyUp, "down": tea.KeyDown}[name]}
}
