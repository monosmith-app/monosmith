package taskfolding

import (
	"reflect"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Fold moves task overrides that two or more projects repeat identically into
// groups. A group selects by language and kind ("lang:node", "kind:lib",
// or both), and is only written when every project it selects that has the
// task either repeats the override or sets its own: a project that would
// start to take the group's values without having asked for them keeps the
// override where it is. A project whose own override differs keeps it, plus
// the group's fields it did not set, so every task resolves as before.
func Fold(projects []Project) Result {
	res := Result{Overrides: map[string]map[string]workspace.TaskOverride{}}
	for _, p := range projects {
		if len(p.Overrides) == 0 {
			continue
		}
		own := make(map[string]workspace.TaskOverride, len(p.Overrides))
		for task, o := range p.Overrides {
			own[task] = o
		}
		res.Overrides[p.Name] = own
	}
	tasks := map[string]bool{}
	for _, p := range projects {
		for task := range p.Overrides {
			tasks[task] = true
		}
	}
	groups := map[string]*workspace.Group{}
	for _, task := range sortedTasks(tasks) {
		best, ok := bestFold(projects, task)
		if !ok {
			continue
		}
		key := strings.Join(best.selector, "\x00")
		g := groups[key]
		if g == nil {
			g = &workspace.Group{Select: best.selector, Tasks: map[string]workspace.TaskOverride{}}
			groups[key] = g
		}
		g.Tasks[task] = best.value
		var names []string
		for _, p := range best.members {
			own := res.Overrides[p.Name]
			if own == nil {
				// An earlier task folded away this name's last override (two
				// projects can share a name): start again, never write to nil.
				own = map[string]workspace.TaskOverride{}
				res.Overrides[p.Name] = own
			}
			if reflect.DeepEqual(own[task], best.value) {
				delete(own, task)
				names = append(names, p.Name)
			} else {
				own[task] = complete(own[task], best.value)
			}
			if len(own) == 0 {
				delete(res.Overrides, p.Name)
			}
		}
		sort.Strings(names)
		res.Folded = append(res.Folded, Folded{Task: task, Select: best.selector, Projects: names})
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		res.Groups = append(res.Groups, *groups[k])
	}
	return res
}

type fold struct {
	selector []string
	value    workspace.TaskOverride
	members  []Project
	repeats  int
}

// bestFold picks the selector that folds the task for the most projects; of
// two that fold as many, the narrower one.
func bestFold(projects []Project, task string) (fold, bool) {
	var best fold
	found := false
	for _, selector := range candidates(projects) {
		var members []Project
		for _, p := range projects {
			if matchesAll(p, selector) && p.Has[task] {
				members = append(members, p)
			}
		}
		value, repeats, ok := repeated(members, task)
		if !ok {
			continue
		}
		if !found || repeats > best.repeats || repeats == best.repeats && len(selector) > len(best.selector) {
			best, found = fold{selector: selector, value: value, members: members, repeats: repeats}, true
		}
	}
	return best, found
}

// repeated is the override two or more members share, when every member has
// an override of its own for the task.
func repeated(members []Project, task string) (workspace.TaskOverride, int, bool) {
	var value workspace.TaskOverride
	best := 0
	for _, m := range members {
		o, ok := m.Overrides[task]
		if !ok {
			return value, 0, false
		}
		n := 0
		for _, other := range members {
			if reflect.DeepEqual(other.Overrides[task], o) {
				n++
			}
		}
		if n > best {
			value, best = o, n
		}
	}
	return value, best, best >= 2
}

// complete gives an override the fields of the group it does not set, so the
// project still resolves to the same task once the group applies to it.
func complete(own, group workspace.TaskOverride) workspace.TaskOverride {
	if own.Command == nil {
		own.Command = group.Command
	}
	if own.Cwd == "" {
		own.Cwd = group.Cwd
	}
	if own.Inputs == nil {
		own.Inputs = group.Inputs
	}
	if own.Outputs == nil {
		own.Outputs = group.Outputs
	}
	if own.DependsOn == nil {
		own.DependsOn = group.DependsOn
	}
	if own.Env == nil {
		own.Env = group.Env
	}
	if own.Cache == nil {
		own.Cache = group.Cache
	}
	if own.Verify == nil {
		own.Verify = group.Verify
	}
	if own.Serve == "" {
		own.Serve = group.Serve
	}
	if own.ServeURL == "" {
		own.ServeURL = group.ServeURL
	}
	if own.ServeTimeout == "" {
		own.ServeTimeout = group.ServeTimeout
	}
	return own
}

// candidates are the selectors a group may use, from the languages and kinds
// the projects have.
func candidates(projects []Project) [][]string {
	seen := map[string]bool{}
	var out [][]string
	add := func(selector ...string) {
		k := strings.Join(selector, "\x00")
		if !seen[k] {
			seen[k] = true
			out = append(out, selector)
		}
	}
	for _, p := range projects {
		if p.Language != "" && p.Kind != "" {
			add("lang:"+p.Language, "kind:"+p.Kind)
		}
		if p.Language != "" {
			add("lang:" + p.Language)
		}
		if p.Kind != "" {
			add("kind:" + p.Kind)
		}
	}
	sort.Slice(out, func(i, j int) bool { return strings.Join(out[i], "\x00") < strings.Join(out[j], "\x00") })
	return out
}

func matchesAll(p Project, selector []string) bool {
	for _, s := range selector {
		switch {
		case strings.HasPrefix(s, "lang:") && s[len("lang:"):] != p.Language:
			return false
		case strings.HasPrefix(s, "kind:") && s[len("kind:"):] != p.Kind:
			return false
		}
	}
	return true
}

func sortedTasks(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
