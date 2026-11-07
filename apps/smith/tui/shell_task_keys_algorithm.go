package tui

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
)

// chooseTaskKeys gives each task a one-letter key. A task takes the first letter
// of its name that is free, else the capital of its first letter, else the first
// free letter of the alphabet. A task that finds none gets no key.
//
// Parameters: tasks are the task names in the order they are drawn; taken says
// whether a key already does something. Returns the key of each task by name.
func chooseTaskKeys(tasks []string, taken func(string) bool) map[string]string {
	given := map[string]bool{}
	keys := map[string]string{}
	free := func(k string) bool { return !given[k] && !taken(k) }
	for _, name := range tasks {
		candidates := strings.Split(strings.ToLower(name), "")
		candidates = append(candidates, strings.ToUpper(name[:1]))
		candidates = append(candidates, strings.Split("abcdefghijklmnopqrstuvwxyz", "")...)
		for _, k := range candidates {
			if len(k) == 1 && strings.ContainsAny(k, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") && free(k) {
				keys[name], given[k] = k, true
				break
			}
		}
	}
	return keys
}

// taskKeys are the keys of the tasks of project p that have no key of their own
// in the keymap, by task name. They are never a key the open view uses.
func (s Shell) taskKeys(p projectlist.Project) map[string]string {
	var others []string
	for _, t := range p.Tasks {
		if !hasOwnKey(t.Name) {
			others = append(others, t.Name)
		}
	}
	return chooseTaskKeys(others, func(k string) bool { return commands.KeyTaken(k, commands.TabProjects) })
}

// hasOwnKey is true for the tasks the keymap binds: build, test, lint, typecheck.
func hasOwnKey(task string) bool {
	for _, id := range []string{"run.build", "run.test", "run.lint", "run.typecheck"} {
		if c, _ := commands.ByID(id); c.Task == task {
			return true
		}
	}
	return false
}
