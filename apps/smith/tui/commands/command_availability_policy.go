package commands

import (
	"slices"
	"unicode"
	"unicode/utf8"
)

// Unavailable says why the command cannot be used now, or "" when it can.
func (c Command) Unavailable(s State) string {
	switch {
	case c.NeedsWorkspace && !s.InWorkspace:
		return "not in a workspace (press " + ShortcutOf("file.new") + " to create one)"
	case c.NeedsRunning && !s.Running:
		return "nothing is running"
	case c.NeedsIdle && s.Running:
		return "a run is in progress (" + ShortcutOf("run.stop") + " stops it)"
	case c.NeedsSelection && !s.HasSelection:
		return "select a task in the navigator first"
	}
	return ""
}

// ownsView is true for the commands that act on the open view's own content:
// they are listed only while that view is open.
func (c Command) ownsView() bool {
	return c.Kind == ViewAction || c.Kind == PaneKey || c.Kind == Follow
}

// inView reports whether the command's keys work in the open view.
func (c Command) inView(s State) bool {
	return len(c.Views) == 0 || slices.Contains(c.Views, s.View)
}

// Visible reports whether the command is listed at all in this state (a
// command for creating a workspace has no place inside one, a view's own
// action none in another view).
func (c Command) Visible(s State) bool {
	if c.Kind == PaneKey {
		return false
	}
	if c.ownsView() && (!c.inView(s) || s.Side) {
		return false
	}
	return (!c.OnlyOutside || !s.InWorkspace) && (!c.OnlyInside || s.InWorkspace) && (!c.NeedsContent || !s.EmptyFolder)
}

// plain is a key that types a character: one letter, digit or symbol.
func plain(key string) bool {
	r, n := utf8.DecodeRuneInString(key)
	return n == len(key) && r != utf8.RuneError && (unicode.IsPrint(r) || r == ' ')
}

// ByKey finds the command a key triggers in the state; an unavailable one is
// returned with ok true and its reason so the shell can say why nothing
// happened. ok is false for an unbound key, for a key the open view handles
// itself (a PaneKey), and for a plain key while a text field or a guided form
// has the keyboard.
func ByKey(key string, s State) (c Command, reason string, ok bool) {
	if plain(key) && (s.Editing || s.Form) {
		return Command{}, "", false
	}
	// The view's own keys win over a global key that is the same letter.
	for _, scoped := range []bool{true, false} {
		for _, cmd := range all {
			if !slices.Contains(cmd.Keys, key) || (len(cmd.Views) > 0) != scoped || !cmd.inView(s) {
				continue
			}
			if cmd.Kind == PaneKey {
				if s.Side {
					continue
				}
				return Command{}, "", false
			}
			if !cmd.Visible(s) || (s.Editing && !cmd.WhileEditing) {
				continue
			}
			return cmd, cmd.Unavailable(s), true
		}
	}
	return Command{}, "", false
}

// ByID finds a command by its id.
func ByID(id string) (Command, bool) {
	for _, c := range all {
		if c.ID == id {
			return c, true
		}
	}
	return Command{}, false
}

// ShortcutOf is how a command's keys are shown, for text that names a key.
func ShortcutOf(id string) string {
	c, _ := ByID(id)
	return c.Shortcut
}
