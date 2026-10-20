package settingsview

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

// Setting is one row of the settings view: a field of smith.toml and its value
// as written (Set is false when it has its default).
type Setting struct {
	Field configedit.Field
	Value string
	Set   bool
}

// Result is what reading smith.toml gave: the shell receives it as a message.
type Result struct {
	Settings []Setting
	Err      error
}

// FromResult is the model for a loaded result.
func FromResult(r Result) Model {
	return Model{Loaded: true, Settings: r.Settings, Err: r.Err}
}

// Model lists smith.toml's settings. Enter edits one (a boolean
// toggles), checked against the schema before it leaves the shell as `smith
// config set`; Ctrl+U (or Delete) returns one to its default with `smith config unset`.
type Model struct {
	Loaded   bool
	Settings []Setting
	Err      error
	Cursor   int
	Editing  bool
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

// Update handles a key; argv is set when the user confirmed a change.
func (v Model) Update(key tea.KeyMsg) (Model, []string) {
	if !v.Loaded || len(v.Settings) == 0 {
		return v, nil
	}
	s := v.Settings[v.Cursor]
	if v.Editing {
		switch key.Type {
		case tea.KeyEsc:
			v.Editing, v.problem = false, ""
			return v, nil
		case tea.KeyEnter:
			values := typed(s.Field, v.input.Value())
			if _, err := s.Field.Encode(values); err != nil {
				v.problem = err.Error()
				return v, nil
			}
			return v, append([]string{"config", "set", s.Field.Path}, values...)
		}
		v.input, _ = v.input.Update(key)
		return v, nil
	}
	switch key.String() {
	case "up":
		v.Cursor = (v.Cursor - 1 + len(v.Settings)) % len(v.Settings)
	case "down":
		v.Cursor = (v.Cursor + 1) % len(v.Settings)
	case "ctrl+u", "delete":
		if s.Set {
			return v, []string{"config", "unset", s.Field.Path}
		}
		v.problem = s.Field.Path + " already has its default"
	case "enter":
		if s.Field.Type == "boolean" {
			return v, []string{"config", "set", s.Field.Path, strconv.FormatBool(s.Value != "true")}
		}
		v.input = textinput.New()
		v.input.SetValue(strings.Join(args(s.Field, s.Value), " "))
		v.input.Focus()
		v.Editing, v.problem = true, ""
	}
	return v, nil
}

// View draws the settings in w cells; rowOf maps a line to the setting drawn on it.
func (v Model) View(th theme.Theme, w int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	if v.Err != nil {
		return []string{th.Failure(v.Err.Error())}, rowOf
	}
	if !v.Loaded {
		return []string{th.Muted("Reading smith.toml…")}, rowOf
	}
	for i, s := range v.Settings {
		value := s.Value
		if !s.Set {
			value = th.Muted("(default)")
		}
		rowOf[len(lines)] = i
		lines = append(lines, guide.Row(th, i == v.Cursor, "", s.Field.Path, value))
	}
	if len(v.Settings) > 0 {
		s := v.Settings[min(v.Cursor, len(v.Settings)-1)]
		lines = append(lines, "", th.Muted(s.Field.Description))
		if v.Editing {
			hint := "one value"
			if s.Field.Type == "array" {
				hint = "values separated by spaces"
				if len(s.Field.Enum) > 0 {
					hint += " (" + strings.Join(s.Field.Enum, ", ") + ")"
				}
			}
			lines = append(lines, "", th.Title(s.Field.Path+" = ")+v.input.View(),
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

// SelectRow moves the cursor to setting i.
func (v Model) SelectRow(i int) Model {
	if v.Loaded && !v.Editing && i >= 0 && i < len(v.Settings) {
		v.Cursor = i
	}
	return v
}
