package doctorview

import (
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
)

// GroupByProject files each finding under the project it is about: the one its
// plugin names (Finding.Project), or, for a finding that names none, the first
// project its check or detail mentions, by name or by directory (a detail says
// "go is not on PATH" and means the tool). Findings about no project (the
// toolchain, git, CI) form the workspace group, which comes first, ordered by
// area; the projects follow in the order given, then any project a finding
// names that the list does not, and only groups with findings appear. Within a
// group the findings keep their order, so problems stay on top.
func GroupByProject(findings []doctor.Finding, projects []ProjectRef) []Group {
	byProject := map[string][]int{}
	var workspace, unlisted []int
	var unlistedNames []string
	listed := map[string]bool{}
	for _, p := range projects {
		listed[p.Name] = true
	}
	for i, f := range findings {
		name := f.Project
		if name == "" {
			name = projectIn(f.Check, f.Detail, projects)
		}
		switch {
		case name == "":
			workspace = append(workspace, i)
		case listed[name]:
			byProject[name] = append(byProject[name], i)
		default:
			if _, seen := byProject[name]; !seen {
				unlistedNames = append(unlistedNames, name)
			}
			byProject[name] = append(byProject[name], i)
			unlisted = append(unlisted, i)
		}
	}
	sort.SliceStable(workspace, func(a, b int) bool {
		return areaRank(findings[workspace[a]].Area) < areaRank(findings[workspace[b]].Area)
	})
	var groups []Group
	if len(workspace) > 0 {
		groups = append(groups, Group{Indexes: workspace})
	}
	for _, p := range projects {
		if idx := byProject[p.Name]; len(idx) > 0 {
			groups = append(groups, Group{Project: p.Name, Indexes: idx})
		}
	}
	for _, name := range unlistedNames {
		groups = append(groups, Group{Project: name, Indexes: byProject[name]})
	}
	return groups
}

// areaRank orders the areas of the workspace group: the tools first, then the
// pipelines, the checks, the dependencies, the release, git and the rest; a
// finding with an area this smith does not know comes last.
func areaRank(area string) int {
	for rank, a := range []string{"toolchain", "ci", "lint", "dependencies", "release", "git", "config"} {
		if a == area {
			return rank
		}
	}
	return 7
}

// projectIn is the project a check and its detail name first, or "".
func projectIn(check, detail string, projects []ProjectRef) string {
	for _, tok := range tokens(check) {
		for _, p := range projects {
			if tok == p.Name || underRoot(tok, p) {
				return p.Name
			}
		}
	}
	for _, tok := range tokens(detail) {
		for _, p := range projects {
			if underRoot(tok, p) {
				return p.Name
			}
		}
	}
	return ""
}

// tokens splits text into words, so "go.work" stays one word and "web." is "web".
func tokens(text string) []string {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return strings.ContainsRune(" \t\n\"'`(),:;", r)
	})
	for i, w := range words {
		words[i] = strings.TrimPrefix(strings.TrimRight(w, "."), "./")
	}
	return words
}

// underRoot reports whether a word is the project's directory or a path inside it.
func underRoot(word string, p ProjectRef) bool {
	return p.Root != "" && p.Root != "." && (word == p.Root || strings.HasPrefix(word, p.Root+"/"))
}

// failuresFirst orders finding indexes in place: failures, then warnings,
// then passes, each keeping its order.
func failuresFirst(findings []doctor.Finding, idx []int) {
	rank := func(f doctor.Finding) int {
		switch {
		case f.OK:
			return 2
		case f.Warning:
			return 1
		}
		return 0
	}
	sort.SliceStable(idx, func(a, b int) bool { return rank(findings[idx[a]]) < rank(findings[idx[b]]) })
}
