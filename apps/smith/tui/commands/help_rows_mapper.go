package commands

// HelpRow is one line of the help: the group it sits under, what it does and
// its keys.
type HelpRow struct {
	Group, Title, Shortcut string
}

// HelpRows lists the keys of the help for a state: first those of the open
// view, then the global ones by group. Both are read from the keymap that the
// keys themselves are, so the help cannot list a key that does nothing.
func HelpRows(s State, viewTitle string) []HelpRow {
	var rows []HelpRow
	for _, c := range all {
		if len(c.Views) > 0 && c.inView(s) && c.Shortcut != "" && (c.Kind == PaneKey || c.Visible(s)) {
			rows = append(rows, HelpRow{Group: viewTitle + " view", Title: c.Title, Shortcut: c.Shortcut})
		}
	}
	for _, g := range GroupOrder {
		for _, c := range all {
			if c.Group == g && len(c.Views) == 0 && c.Shortcut != "" && (c.Kind == PaneKey || c.Visible(s)) {
				rows = append(rows, HelpRow{Group: g, Title: c.Title, Shortcut: c.Shortcut})
			}
		}
	}
	return rows
}
