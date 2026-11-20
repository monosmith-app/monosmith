package release

import "slices"

// PlannedItems picks the items of a plan whose project is one of projects, in
// the plan's order (dependency order). A project the plan does not release
// has no item, so it keeps the version of its manifest.
//
// Parameters: plan is the release plan; projects are the project names to
// look for.
// Returns: the matching items, or nil when none match.
func PlannedItems(plan Plan, projects []string) []Item {
	var out []Item
	for _, item := range plan.Items {
		if slices.Contains(projects, item.Project) {
			out = append(out, item)
		}
	}
	return out
}
