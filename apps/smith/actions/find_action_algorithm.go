package actions

// Find returns the action a command line runs: the one with the longest
// command path that begins the line. ok is false when no action does.
func Find(argv []string) (action Action, ok bool) {
	best := -1
	for _, a := range catalog {
		if len(a.Command) <= best || len(a.Command) > len(argv) {
			continue
		}
		matches := true
		for i, w := range a.Command {
			if argv[i] != w {
				matches = false
				break
			}
		}
		if matches {
			action, best, ok = a, len(a.Command), true
		}
	}
	return action, ok
}

// All returns every action.
func All() []Action { return append([]Action(nil), catalog...) }

// ByID returns the action named id.
func ByID(id string) (Action, bool) {
	for _, a := range catalog {
		if a.ID() == id {
			return a, true
		}
	}
	return Action{}, false
}
