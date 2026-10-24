package pluginsview

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Model lists the plugins, what each fills and which projects use it.
// Turning one on or off ends as `smith plugin enable|disable <name>`.
type Model struct {
	Loaded bool
	Rows   []Row
	Err    error
	Cursor int
}

// FromResult is the model for a loaded result; the cursor stays where it was
// when that row still exists.
func (v Model) FromResult(r Result) Model {
	cursor := v.Cursor
	if cursor >= len(r.Rows) {
		cursor = 0
	}
	return Model{Loaded: true, Rows: r.Rows, Err: r.Err, Cursor: cursor}
}

// Reset forgets what was loaded but keeps the cursor.
func (v Model) Reset() Model { return Model{Cursor: v.Cursor} }

// Update handles a key; page is how many rows a page key moves.
func (v Model) Update(key tea.KeyMsg, page int) Model {
	v.Cursor = pane.MoveCursor(v.Cursor, len(v.Rows), key.String(), page)
	return v
}

// current is the row under the cursor.
func (v Model) current() (Row, bool) {
	if v.Cursor < 0 || v.Cursor >= len(v.Rows) {
		return Row{}, false
	}
	return v.Rows[v.Cursor], true
}

// ToggleArgv is the command line that turns the current plugin on or off; nil
// when it cannot be switched.
func (v Model) ToggleArgv() []string {
	r, ok := v.current()
	if !ok || !r.Switchable {
		return nil
	}
	verb := "enable"
	if r.Enabled {
		verb = "disable"
	}
	return []string{"plugin", verb, r.Name}
}

// SelectRow moves the cursor to row i.
func (v Model) SelectRow(i int) Model {
	if i >= 0 && i < len(v.Rows) {
		v.Cursor = i
	}
	return v
}

// View draws the list in a w by h area; rowOf maps a line to the row drawn on it.
func (v Model) View(th theme.Theme, w, h int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	switch {
	case v.Err != nil:
		return []string{th.Failure(v.Err.Error())}, rowOf
	case !v.Loaded:
		return []string{th.Muted("Listing the plugins…")}, rowOf
	case len(v.Rows) == 0:
		return []string{th.Muted("No plugins found.")}, rowOf
	}
	builtin, installed, on := 0, 0, 0
	for _, r := range v.Rows {
		if r.Builtin {
			builtin++
		} else {
			installed++
		}
		if r.Enabled {
			on++
		}
	}
	inUse := 0
	for _, r := range v.Rows {
		if len(r.UsedBy) > 0 {
			inUse++
		}
	}
	summary := fmt.Sprintf("%d built-in · %d installed · %d on", builtin, installed, on)
	if inUse > 0 {
		summary += fmt.Sprintf(" · %d used by your projects (listed first)", inUse)
	} else {
		summary += " · none used by a project yet"
	}
	lines = append(lines, pane.Wrap(th.Muted(summary), w, 0)...)
	lines = append(lines, "")
	nameW := 8
	for _, r := range v.Rows {
		nameW = max(nameW, pane.Width(r.Name))
	}
	nameW = min(nameW+1, max(w/4, 8))
	const stateW = 11
	rest := max(w-2-nameW-stateW, 12)
	rolesW := rest * 11 / 20
	usedW := rest - rolesW
	row := func(name, state, roles, used string) string {
		return "  " + pane.Cell(name, nameW) + pane.Cell(state, stateW) + pane.Cell(roles, rolesW) + pane.Cell(used, usedW)
	}
	lines = append(lines, th.Muted(pane.Fit(row("Plugin", "State", "Roles", "Used by"), w)))

	detail := v.detail(th, w)
	avail := max(h-len(lines)-len(detail), 1)
	top := pane.WindowTop(v.Cursor, len(v.Rows), avail)
	for i := top; i < len(v.Rows) && i < top+avail; i++ {
		r := v.Rows[i]
		state := th.Muted("off")
		switch {
		case !r.Builtin:
			state = th.Info("installed")
		case r.Enabled:
			state = th.Success("on")
		}
		used := "-"
		if len(r.UsedBy) > 0 {
			used = strings.Join(r.UsedBy, ", ")
		}
		line := row(r.Name, state, strings.Join(r.Roles, ", "), used)
		if i == v.Cursor {
			line = th.Selected(pane.Fit(pane.Strip(line), w))
		}
		rowOf[len(lines)] = i
		lines = append(lines, line)
	}
	lines = append(lines, detail...)
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, rowOf
}

// detail describes the current plugin below the list.
func (v Model) detail(th theme.Theme, w int) []string {
	r, ok := v.current()
	if !ok {
		return nil
	}
	head := th.Bold(r.Name)
	if r.Version != "" {
		head += " " + th.Muted(r.Version)
	}
	out := []string{"", th.Faint(strings.Repeat("─", w)), head, "  " + th.Muted("from     ") + r.Source}
	if len(r.Roles) > 0 {
		out = append(out, pane.Wrap("  "+th.Muted("roles    ")+strings.Join(r.Roles, ", "), w, 11)...)
	}
	if len(r.Requires) > 0 {
		out = append(out, "  "+th.Muted("needs    ")+strings.Join(r.Requires, ", "))
	}
	switch {
	case len(r.UsedBy) > 0:
		out = append(out, pane.Wrap("  "+th.Muted("used by  ")+strings.Join(r.UsedBy, ", "), w, 11)...)
	case r.Enabled:
		out = append(out, "  "+th.Muted("used by  no project"))
	}
	if argv := v.ToggleArgv(); argv != nil {
		verb := "enable"
		if r.Enabled {
			verb = "disable"
		}
		out = append(out, pane.Wrap("  "+th.Muted("Press e to "+verb+", or run: "+commands.Line(argv)), w, 4)...)
	}
	return out
}
