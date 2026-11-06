package settingsview

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pathfinder"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
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
	// editSearch asks what to look for in the plugin marketplace.
	editSearch
)

// editor is what fills the right pane while a value is being changed.
type editor struct {
	kind  editorKind
	title string
	// about is what the thing does, shown above the field.
	about   string
	setting Setting
	input   textinput.Model
	// finder, when set, lists the folders and files under the field: the value
	// is a path (a setting such as cache.dir, or the source of a plugin).
	finder *pathfinder.Model
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
		placeholder = words.Say("settingsview.editor.placeholder", "type", st.Type)
	}
	e := &editor{kind: editText, title: st.Path, about: st.Explanation, setting: st, input: newInput(value, placeholder)}
	if takesPath(st) {
		e.withFinder()
	}
	return e
}

// withFinder puts the list of folders and files under the field.
func (e *editor) withFinder() {
	f := pathfinder.New("", nil).Sync(e.input.Value())
	e.finder = &f
}

// takesPath reports whether a setting holds a path, by the name of its key:
// cache.dir, workspace.recipe.
func takesPath(st Setting) bool {
	if st.Type != "string" {
		return false
	}
	key := st.Path[strings.LastIndex(st.Path, ".")+1:]
	if key == "recipe" {
		return true
	}
	for _, suffix := range []string{"dir", "file", "path", "folder"} {
		if strings.HasSuffix(key, suffix) {
			return true
		}
	}
	return false
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
	return &editor{kind: editReview, title: words.Say("settingsview.editor.trust.title", "name", name), plugin: name, text: r.Text, trusted: r.Trusted}
}

func newInstall() *editor {
	e := &editor{
		kind:  editInstall,
		title: words.Say("settingsview.install.title"),
		about: words.Say("settingsview.install.about"),
		input: newInput("", words.Say("settingsview.install.placeholder")),
	}
	e.withFinder()
	return e
}

func newSearch() *editor {
	return &editor{
		kind:  editSearch,
		title: words.Say("settingsview.search.title"),
		about: words.Say("settingsview.search.about"),
		input: newInput("", words.Say("settingsview.search.placeholder")),
	}
}

// installArgv is the command line the typed text stands for, or why it is not one.
func installArgv(typed string) ([]string, string) {
	fields := strings.Fields(typed)
	if len(fields) == 0 {
		return []string{"plugin", "install", "<source>"}, words.Say("settingsview.install.missing")
	}
	argv := []string{"plugin", "install"}
	if n := len(fields); n > 1 && isSHA256(fields[n-1]) {
		argv = append(argv, strings.Join(fields[:n-1], " "), "--sha256", fields[n-1])
	} else {
		argv = append(argv, strings.Join(fields, " "))
	}
	return argv, ""
}

func isSHA256(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
			return false
		}
	}
	return true
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
	case editSearch:
		if name == "enter" {
			m.ed = nil
			return m, Result{Search: true, Term: strings.TrimSpace(e.input.Value())}
		}
		e.input, _ = e.input.Update(msg)
	case editText, editInstall:
		if e.finder != nil {
			f, out := e.finder.Handle(msg)
			e.finder = &f
			if out.Handled {
				if out.Set {
					e.input.SetValue(out.Value)
					e.input.CursorEnd()
					e.err = ""
				}
				if !out.Confirm {
					m.ed = &e
					return m, Result{}
				}
				name = "enter"
			}
		}
		if name == "enter" {
			e.err = ""
			m.ed = &e
			if e.kind == editInstall {
				argv, why := installArgv(e.input.Value())
				if why != "" {
					e.err = why
					m.ed = &e
					return m, Result{}
				}
				m.ed = nil
				return m, Result{Run: argv}
			}
			m = m.begin(words.Say("settingsview.busy.saving", "path", e.setting.Path), false)
			return m, Result{Cmd: m.setSetting(e.setting.Path, e.values())}
		}
		e.err = ""
		e.input, _ = e.input.Update(msg)
		if e.finder != nil {
			f := e.finder.Sync(e.input.Value())
			e.finder = &f
		}
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
			m = m.begin(words.Say("settingsview.busy.saving", "path", e.setting.Path), false)
			return m, Result{Cmd: m.setSetting(e.setting.Path, e.values())}
		}
	case editReview:
		switch name {
		case "y":
			m.ed = &e
			m = m.begin(words.Say("settingsview.busy.trusting", "name", e.plugin), false)
			return m, Result{Cmd: m.trust(e.plugin, false)}
		case "p":
			m.ed = &e
			m = m.begin(words.Say("settingsview.busy.pinning", "name", e.plugin), false)
			return m, Result{Cmd: m.trust(e.plugin, true)}
		case "f", "u":
			if e.trusted {
				m.ed = &e
				m = m.begin(words.Say("settingsview.busy.forgetting", "name", e.plugin), false)
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
