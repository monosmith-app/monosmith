package commands

import "github.com/monosmith-app/monosmith/libs/words/words"

// HelpRow is one line of the help: the group it sits under, what it does and
// its keys.
type HelpRow struct {
	Group, Title, Shortcut string
	// Disabled is why the key cannot run now, or "" when it can.
	Disabled string
}

// HelpRows lists the keys of the help for a state: first those of the open
// view, then the global ones by group. Both are read from the keymap that the
// keys themselves are, so the help cannot list a key that does nothing. A key
// that cannot run now carries its reason (Disabled); outside a workspace, where
// most keys have nothing to act on, only the keys that can run are listed.
func HelpRows(s State, viewTitle string) []HelpRow {
	var rows []HelpRow
	add := func(group string, c Command) {
		why := ""
		if c.Kind != PaneKey || s.Reasons[c.ID] != "" {
			why = c.Unavailable(s)
		}
		if why != "" && !s.InWorkspace {
			return
		}
		rows = append(rows, HelpRow{Group: group, Title: c.Title, Shortcut: c.Shortcut, Disabled: why})
	}
	for _, c := range all {
		if len(c.Views) > 0 && c.inView(s) && c.Shortcut != "" && c.gated(s) && (c.Kind == PaneKey || c.Visible(s)) {
			add(words.Say("commands.help.view", "view", viewTitle), c)
		}
	}
	for _, g := range GroupOrder {
		for _, c := range all {
			if c.Group == g && len(c.Views) == 0 && c.Shortcut != "" && c.gated(s) && (c.Kind == PaneKey || c.Visible(s)) {
				add(g, c)
			}
		}
	}
	return rows
}
