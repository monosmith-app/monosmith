package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// pluginRow is one plugin as the Plugins view shows it.
type pluginRow struct {
	Name    string
	Version string
	// Source is "built-in" or the directory an installed plugin was found in.
	Source  string
	Builtin bool
	// Switchable is true for a built-in plugin that smith.toml can turn on or off.
	Switchable bool
	Enabled    bool
	// Roles are the levels (language, deploy) and slots (lint) it fills.
	Roles    []string
	Requires []string
	// UsedBy are the projects that plugin discovered or deploys.
	UsedBy []string
}

type pluginsLoaded struct {
	rows []pluginRow
	err  error
}

// pluginsView lists the plugins, what each fills and which projects use it.
// Turning one on or off ends as `smith plugin enable|disable <name>`.
type pluginsView struct {
	loaded bool
	rows   []pluginRow
	err    error
	cursor int
}

func (v pluginsView) update(key tea.KeyMsg, page int) pluginsView {
	v.cursor = moveCursor(v.cursor, len(v.rows), key.String(), page)
	return v
}

// current is the row under the cursor.
func (v pluginsView) current() (pluginRow, bool) {
	if v.cursor < 0 || v.cursor >= len(v.rows) {
		return pluginRow{}, false
	}
	return v.rows[v.cursor], true
}

// toggleArgv is the command line that turns the current plugin on or off; nil
// when it cannot be switched.
func (v pluginsView) toggleArgv() []string {
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

// selectRow moves the cursor to row i.
func (v pluginsView) selectRow(i int) pluginsView {
	if i >= 0 && i < len(v.rows) {
		v.cursor = i
	}
	return v
}

// view draws the list in a w by h area; rowOf maps a line to the row drawn on it.
func (v pluginsView) view(th theme.Theme, w, h int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	switch {
	case v.err != nil:
		return []string{th.Failure(v.err.Error())}, rowOf
	case !v.loaded:
		return []string{th.Muted("Listing the plugins…")}, rowOf
	case len(v.rows) == 0:
		return []string{th.Muted("No plugins found.")}, rowOf
	}
	builtin, installed, on := 0, 0, 0
	for _, r := range v.rows {
		if r.Builtin {
			builtin++
		} else {
			installed++
		}
		if r.Enabled {
			on++
		}
	}
	lines = append(lines, th.Muted(fmt.Sprintf("%d built-in · %d installed · %d on", builtin, installed, on)), "")
	nameW := 8
	for _, r := range v.rows {
		nameW = max(nameW, pane.Width(r.Name))
	}
	nameW = min(nameW, max(w/4, 8))
	lines = append(lines, th.Muted(pane.Fit("  "+pad("Plugin", nameW)+" "+pad("State", 10)+pad("Roles", 24)+"Used by", w)))

	detail := v.detail(th, w)
	avail := max(h-len(lines)-len(detail), 1)
	top := windowTop(v.cursor, len(v.rows), avail)
	for i := top; i < len(v.rows) && i < top+avail; i++ {
		r := v.rows[i]
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
		line := "  " + pad(r.Name, nameW) + " " + pad(state, 10) + pad(strings.Join(r.Roles, ", "), 24) + used
		if i == v.cursor {
			line = th.Selected(pane.Fit(pane.Strip("  "+pad(r.Name, nameW)+" "+pad(state, 10)+pad(strings.Join(r.Roles, ", "), 24)+used), w))
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
func (v pluginsView) detail(th theme.Theme, w int) []string {
	r, ok := v.current()
	if !ok {
		return nil
	}
	head := th.Bold(r.Name)
	if r.Version != "" {
		head += " " + th.Muted(r.Version)
	}
	out := []string{"", th.Faint(strings.Repeat("─", w)), head, "  " + th.Muted("from     ") + r.Source}
	if len(r.Requires) > 0 {
		out = append(out, "  "+th.Muted("needs    ")+strings.Join(r.Requires, ", "))
	}
	switch {
	case len(r.UsedBy) > 0:
		out = append(out, "  "+th.Muted("used by  ")+strings.Join(r.UsedBy, ", "))
	case r.Enabled:
		out = append(out, "  "+th.Muted("used by  no project"))
	}
	if argv := v.toggleArgv(); argv != nil {
		out = append(out, "  "+th.Muted("e runs smith "+strings.Join(argv, " ")))
	}
	return out
}
