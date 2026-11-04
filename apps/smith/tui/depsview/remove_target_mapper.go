package depsview

import "slices"

// rootLabel is how the Deps view names the manifest at the workspace root.
const rootLabel = "(root)"

// rootTarget is the -w value that smith uninstall takes for the workspace root.
const rootTarget = "."

// Target is a project a dependency is declared in, as smith uninstall -w names it.
type Target struct {
	// Label is what the view shows: the project name, or "(root)".
	Label string
	// Value is the -w value: the project name, or "." for the workspace root.
	Value string
}

// RemoveTargets are the projects that declare the package under the cursor,
// each once, in the order the view lists them.
//
// Returns: the package name, its declaring projects, and false when there is
// no row under the cursor.
func (v Model) RemoveTargets() (name string, targets []Target, ok bool) {
	if v.Cursor < 0 || v.Cursor >= len(v.Rows) {
		return "", nil, false
	}
	row := v.Rows[v.Cursor]
	var seen []string
	for _, p := range row.Projects {
		if slices.Contains(seen, p) {
			continue
		}
		seen = append(seen, p)
		t := Target{Label: p, Value: p}
		if p == rootLabel {
			t.Value = rootTarget
		}
		targets = append(targets, t)
	}
	return row.Name, targets, len(targets) > 0
}
