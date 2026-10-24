package settingsview

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
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
		v.input.Prompt = ""
		v.input.SetValue(strings.Join(args(s.Field, s.Value), " "))
		v.input.Focus()
		v.Editing, v.problem = true, ""
	}
	return v, nil
}

// View draws the settings in a w by h area: the list on top and, right under
// it, what the selected setting does, what it accepts and what it is now.
// Editing happens on the selected row itself. rowOf maps a line to the
// setting drawn on it.
func (v Model) View(th theme.Theme, w, h int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	if v.Err != nil {
		return []string{th.Failure(v.Err.Error())}, rowOf
	}
	if !v.Loaded {
		return []string{th.Muted("Reading smith.toml…")}, rowOf
	}
	if len(v.Settings) == 0 {
		return []string{th.Muted("No settings.")}, rowOf
	}
	cur := min(v.Cursor, len(v.Settings)-1)
	detail := v.detail(th, w, cur)
	lines = append(lines, pane.Wrap(th.Muted("Settings in smith.toml. One you have not set shows its default."), w, 0)...)
	lines = append(lines, "")
	pathW := 0
	for _, s := range v.Settings {
		pathW = max(pathW, pane.Width(s.Field.Path))
	}
	pathW = min(pathW+2, max(w/2, 12))
	valueW := max(w-2-pathW, 8)

	avail := max(h-len(lines)-len(detail), 3)
	top := pane.WindowTop(cur, len(v.Settings), avail)
	for i := top; i < len(v.Settings) && i < top+avail; i++ {
		s := v.Settings[i]
		value := s.Value
		switch {
		case i == cur && v.Editing:
			in := v.input
			in.Width = max(valueW-3, 4)
			value = "[" + in.View() + "]"
		case !s.Set && s.Field.Default() != "":
			value = th.Muted("default: " + s.Field.Default())
		case !s.Set:
			value = th.Muted("(default)")
		}
		rowOf[len(lines)] = i
		cursor, path := "  ", pane.Cell(s.Field.Path, pathW)
		if i == cur {
			cursor, path = th.Accent(th.Icons.Pointer+" "), th.Accent(path)
		}
		lines = append(lines, cursor+path+pane.Cut(value, valueW))
	}
	lines = append(lines, detail...)
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, rowOf
}

// detail describes the setting at cur below the list.
func (v Model) detail(th theme.Theme, w, cur int) []string {
	s := v.Settings[cur]
	out := []string{"", th.Faint(strings.Repeat("─", w)), th.Bold(s.Field.Path)}
	out = append(out, pane.Wrap("  "+s.Field.Explanation(), w, 2)...)
	out = append(out, "  "+th.Muted("accepts  ")+allowed(s.Field))
	now := "its default"
	if s.Set {
		now = s.Value
	} else if d := s.Field.Default(); d != "" {
		now = "its default, " + d
	}
	out = append(out, "  "+th.Muted("now      ")+now)
	if v.Editing {
		out = append(out, pane.Wrap(th.Muted("Enter saves (smith config set) · Esc cancels"), w, 0)...)
	}
	if v.problem != "" {
		out = append(out, th.Failure(v.problem))
	}
	return out
}

// SelectRow moves the cursor to setting i.
func (v Model) SelectRow(i int) Model {
	if v.Loaded && !v.Editing && i >= 0 && i < len(v.Settings) {
		v.Cursor = i
	}
	return v
}
