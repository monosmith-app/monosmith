package release

import (
	"fmt"
	"slices"
	"strings"
)

// Adjustments are what the person changed in a release plan: packages held
// back and bumps they chose themselves.
type Adjustments struct {
	// Hold names the packages that do not release this time.
	Hold []string
	// Bump maps a package to "patch", "minor" or "major", replacing the level
	// its commits asked for.
	Bump map[string]string
}

// Adjust applies the person's choices to a plan. A held package leaves the
// plan, and so does a package that releases only because a held one does. A
// bump replaces the level of one package: its next version and tag follow.
//
// Parameters: plan is what the commits asked for; a is what the person chose.
// Returns: the adjusted plan and one note per change (shown to the person); an
// error for a name that is not in the plan, a level smith does not know, or a
// package whose version the commits do not decide (first release, forced by a
// version() directive, or a channel release).
func Adjust(plan Plan, a Adjustments) (Plan, []string, error) {
	out := plan
	out.Items = slices.Clone(plan.Items)
	var notes []string
	in := func(name string) int {
		return slices.IndexFunc(out.Items, func(i Item) bool { return i.Project == name })
	}
	for _, name := range sortedKeys(a.Bump) {
		i := in(name)
		if i < 0 {
			return plan, nil, unknownInPlan(plan, "--bump", name)
		}
		item := out.Items[i]
		level, ok := levelNamed(a.Bump[name])
		if !ok || level == None {
			return plan, nil, fmt.Errorf("--bump %s=%s: the level is patch, minor or major", name, a.Bump[name])
		}
		switch {
		case plan.Channel != "":
			return plan, nil, fmt.Errorf("--bump %s: a channel release has its own counter, so the bump cannot change", name)
		case item.Forced:
			return plan, nil, fmt.Errorf("--bump %s: a version(%s)[...] commit already sets the version", name, name)
		case item.Current == "":
			return plan, nil, fmt.Errorf("--bump %s: %s was never released, so there is no version to bump; set it with a version(%s)[x.y.z] commit", name, name, name)
		}
		current, err := ParseVersion(item.Current)
		if err != nil {
			return plan, nil, fmt.Errorf("--bump %s: %w", name, err)
		}
		next := current.Bump(level)
		if next.String() == item.Next {
			continue
		}
		asked := item.Level
		item.Tag = strings.TrimSuffix(item.Tag, item.Next) + next.String()
		item.Level, item.Next = level.String(), next.String()
		item.Reason = fmt.Sprintf("you chose %s; the commits asked for %s", level, asked)
		out.Items[i] = item
		notes = append(notes, fmt.Sprintf("%s: %s bump, %s -> %s (the commits asked for %s)", name, level, item.Current, item.Next, asked))
	}
	held := map[string]bool{}
	for _, name := range a.Hold {
		if in(name) < 0 {
			if slices.Contains(plan.Unchanged, name) {
				continue
			}
			return plan, nil, unknownInPlan(plan, "--hold", name)
		}
		held[name] = true
	}
	for changed := true; changed; {
		changed = false
		for _, item := range out.Items {
			if held[item.Project] {
				continue
			}
			for h := range held {
				if strings.HasPrefix(item.Reason, "dependency "+h+" ") {
					held[item.Project] = true
					changed = true
				}
			}
		}
	}
	var kept []Item
	for _, item := range out.Items {
		if !held[item.Project] {
			kept = append(kept, item)
			continue
		}
		why := "held back by you"
		if !slices.Contains(a.Hold, item.Project) {
			why = "held back because a package it depends on is held back"
		}
		notes = append(notes, fmt.Sprintf("%s: %s (it stays at %s)", item.Project, why, orUnreleased(item.Current)))
	}
	out.Items = kept
	return out, notes, nil
}

func orUnreleased(v string) string {
	if v == "" {
		return "unreleased"
	}
	return v
}

func unknownInPlan(plan Plan, flag, name string) error {
	var names []string
	for _, i := range plan.Items {
		names = append(names, i.Project)
	}
	return fmt.Errorf("%s %s: %s has nothing to release in this plan (the plan holds: %s)", flag, name, name, strings.Join(names, ", "))
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// ParseBumps turns the values of --bump (project=level) into a map.
//
// Parameters: values are the flag values.
// Returns: the map; an error for a value without "=" or a level smith does not know.
func ParseBumps(values []string) (map[string]string, error) {
	out := map[string]string{}
	for _, v := range values {
		name, level, ok := strings.Cut(v, "=")
		if !ok || name == "" {
			return nil, fmt.Errorf("--bump %q: write it as <project>=patch|minor|major", v)
		}
		if l, known := levelNamed(level); !known || l == None {
			return nil, fmt.Errorf("--bump %q: the level is patch, minor or major", v)
		}
		out[name] = level
	}
	return out, nil
}

// Consequences says in plain words what a confirmed release does and that it
// cannot be undone.
//
// Parameters: plan is the release; push is false with --no-push.
// Returns: one sentence.
func Consequences(plan Plan, push bool) string {
	tags := make([]string, 0, len(plan.Items))
	for _, i := range plan.Items {
		tags = append(tags, i.Tag)
	}
	where := "and push them to the remote"
	if !push {
		where = "and keep them on this machine (--no-push)"
	}
	return fmt.Sprintf("This cannot be undone. smith will publish %d package(s) to their registries, create the tags %s %s.",
		len(plan.Items), strings.Join(tags, ", "), where)
}
