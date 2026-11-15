package legacyinventory

import (
	"sort"
	"strings"
)

// ToolsInUse names the tooling the workspace has today that the full level
// can retire or replace, for a person to read: "Nx", "mnci", "ESLint",
// "Jest", "Vitest", then the build tools (rollup, webpack, ...) sorted. It
// says what is there, not what would be removed: that depends on the choices
// and on what still uses each tool.
func ToolsInUse(t *Tree) []string {
	usage := Detect(t)
	nx, mnci := false, false
	builds := map[string]bool{}
	note := func(name string, family Family) {
		switch family {
		case FamilyNx:
			if strings.Contains(name, "mnci") {
				mnci = true
			} else {
				nx = true
			}
		case FamilyBuild:
			builds[buildToolName(name)] = true
		}
	}
	for _, f := range t.Files() {
		if r, ok := fileFamily(f); ok {
			if r.family == FamilyBuild {
				note(r.tool, r.family)
			} else {
				note(f, r.family)
			}
		}
	}
	for _, mf := range t.ManifestPaths() {
		m, ok := ReadManifest(t.Read(mf))
		if !ok {
			continue
		}
		for _, name := range m.Declared() {
			if family, ok := packageFamily(name); ok {
				note(name, family)
			}
		}
	}
	var out []string
	for _, tool := range []struct {
		name string
		on   bool
	}{{"Nx", nx}, {"mnci", mnci}, {"ESLint", usage.ESLint}, {"Jest", usage.Jest}, {"Vitest", usage.Vitest}} {
		if tool.on {
			out = append(out, tool.name)
		}
	}
	var rest []string
	for name := range builds {
		rest = append(rest, name)
	}
	sort.Strings(rest)
	return append(out, rest...)
}

// buildToolName is the tool a build file or package belongs to.
func buildToolName(name string) string {
	switch {
	case strings.Contains(name, "rollup"):
		return "rollup"
	case strings.Contains(name, "webpack"), strings.HasSuffix(name, "-loader"):
		return "webpack"
	case strings.Contains(name, "swc"):
		return "swc"
	case strings.Contains(name, "babel"):
		return "babel"
	}
	return name
}
