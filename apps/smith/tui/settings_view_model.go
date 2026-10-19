package tui

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// setting is one row of the settings view: a field of smith.toml and its value
// as written (set is false when it has its default).
type setting struct {
	field configedit.Field
	value string
	set   bool
}

type settingsLoaded struct {
	settings []setting
	err      error
}

// settingsView lists smith.toml's settings. Enter edits one (a boolean
// toggles), checked against the schema before it leaves the shell as `smith
// config set`; Ctrl+U (or Delete) returns one to its default with `smith config unset`.
type settingsView struct {
	loaded   bool
	settings []setting
	err      error
	cursor   int
	editing  bool
	input    textinput.Model
	problem  string
}

var quoted = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)

// args are a value as written, as the arguments `smith config set` takes.
func args(f configedit.Field, value string) []string {
	switch f.Type {
	case "array":
		var out []string
		for _, m := range quoted.FindAllStringSubmatch(value, -1) {
			out = append(out, m[1])
		}
		return out
	case "string":
		if s, err := strconv.Unquote(value); err == nil {
			return []string{s}
		}
	}
	if value == "" {
		return nil
	}
	return []string{value}
}

// typed splits what was typed into arguments: a list takes words, anything
// else is one value.
func typed(f configedit.Field, text string) []string {
	if f.Type == "array" {
		return strings.Fields(text)
	}
	return []string{strings.TrimSpace(text)}
}

// update handles a key; argv is set when the user confirmed a change.
func (v settingsView) update(key tea.KeyMsg) (settingsView, []string) {
	if !v.loaded || len(v.settings) == 0 {
		return v, nil
	}
	s := v.settings[v.cursor]
	if v.editing {
		switch key.Type {
		case tea.KeyEsc:
			v.editing, v.problem = false, ""
			return v, nil
		case tea.KeyEnter:
			values := typed(s.field, v.input.Value())
			if _, err := s.field.Encode(values); err != nil {
				v.problem = err.Error()
				return v, nil
			}
			return v, append([]string{"config", "set", s.field.Path}, values...)
		}
		v.input, _ = v.input.Update(key)
		return v, nil
	}
	switch key.String() {
	case "up":
		v.cursor = (v.cursor - 1 + len(v.settings)) % len(v.settings)
	case "down":
		v.cursor = (v.cursor + 1) % len(v.settings)
	case "ctrl+u", "delete":
		if s.set {
			return v, []string{"config", "unset", s.field.Path}
		}
		v.problem = s.field.Path + " already has its default"
	case "enter":
		if s.field.Type == "boolean" {
			return v, []string{"config", "set", s.field.Path, strconv.FormatBool(s.value != "true")}
		}
		v.input = textinput.New()
		v.input.SetValue(strings.Join(args(s.field, s.value), " "))
		v.input.Focus()
		v.editing, v.problem = true, ""
	}
	return v, nil
}

// view draws the settings in w cells; rowOf maps a line to the setting drawn on it.
func (v settingsView) view(th theme.Theme, w int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	if v.err != nil {
		return []string{th.Failure(v.err.Error())}, rowOf
	}
	if !v.loaded {
		return []string{th.Muted("Reading smith.toml…")}, rowOf
	}
	for i, s := range v.settings {
		value := s.value
		if !s.set {
			value = th.Muted("(default)")
		}
		rowOf[len(lines)] = i
		lines = append(lines, guide.Row(th, i == v.cursor, "", s.field.Path, value))
	}
	if len(v.settings) > 0 {
		s := v.settings[min(v.cursor, len(v.settings)-1)]
		lines = append(lines, "", th.Muted(s.field.Description))
		if v.editing {
			hint := "one value"
			if s.field.Type == "array" {
				hint = "values separated by spaces"
				if len(s.field.Enum) > 0 {
					hint += " (" + strings.Join(s.field.Enum, ", ") + ")"
				}
			}
			lines = append(lines, "", th.Title(s.field.Path+" = ")+v.input.View(),
				th.Muted(hint+" · enter save (smith config set) · esc cancel"))
		} else {
			lines = append(lines, "", th.Muted("up/down move · enter edit (a yes/no setting flips) · u back to the default"))
		}
	}
	if v.problem != "" {
		lines = append(lines, th.Failure(v.problem))
	}
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, rowOf
}

// selectRow moves the cursor to setting i.
func (v settingsView) selectRow(i int) settingsView {
	if v.loaded && !v.editing && i >= 0 && i < len(v.settings) {
		v.cursor = i
	}
	return v
}
