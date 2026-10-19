package commands

import "github.com/monosmith-app/monosmith/apps/smith/tui/menubar"

// Menus builds the menu bar's menus for a state: every visible command in its
// menu with its shortcut, and the reason beside those that are unavailable.
func Menus(s State) []menubar.Menu {
	byMenu := map[string][]menubar.Entry{}
	for _, c := range all {
		if c.Menu == "" || !c.Visible(s) {
			continue
		}
		if separatorBefore[c.ID] {
			byMenu[c.Menu] = append(byMenu[c.Menu], menubar.Entry{Separator: true})
		}
		byMenu[c.Menu] = append(byMenu[c.Menu], menubar.Entry{
			ID: c.ID, Name: c.Title, Shortcut: c.Shortcut, Hint: c.Hint, Disabled: c.Unavailable(s),
		})
	}
	var menus []menubar.Menu
	for _, name := range MenuOrder {
		if entries := byMenu[name]; len(entries) > 0 {
			menus = append(menus, menubar.Menu{Name: name, Mnemonic: Mnemonics[name], Entries: entries})
		}
	}
	return menus
}

// Shortcuts lists the visible commands that have a shortcut, for the help
// overlay: each with its menu, title and key.
func Shortcuts(s State) []Command {
	var out []Command
	for _, c := range all {
		if c.Shortcut != "" && c.Visible(s) {
			out = append(out, c)
		}
	}
	return out
}
