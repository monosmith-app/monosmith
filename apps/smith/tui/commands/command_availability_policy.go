package commands

// Unavailable says why the command cannot be used now, or "" when it can.
func (c Command) Unavailable(s State) string {
	switch {
	case c.NeedsWorkspace && !s.InWorkspace:
		return "not in a workspace (create one with File > New workspace)"
	case c.NeedsRunning && !s.Running:
		return "nothing is running"
	case c.NeedsIdle && s.Running:
		return "a run is in progress (Ctrl+C stops it)"
	case c.NeedsSelection && !s.HasSelection:
		return "select a project or task in the navigator first"
	}
	return ""
}

// Visible reports whether the command is listed at all in this state (a
// command for creating a workspace has no place inside one).
func (c Command) Visible(s State) bool {
	return !(c.OnlyOutside && s.InWorkspace) && !(c.OnlyInside && !s.InWorkspace)
}

// ByKey finds the command a key triggers among those visible and available
// in the state; an unavailable one is returned with ok true and its reason so
// the shell can say why nothing happened. ok is false for an unbound key.
func ByKey(key string, s State) (c Command, reason string, ok bool) {
	for _, cmd := range all {
		if !cmd.Visible(s) {
			continue
		}
		for _, k := range cmd.Keys {
			if k == key {
				return cmd, cmd.Unavailable(s), true
			}
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
