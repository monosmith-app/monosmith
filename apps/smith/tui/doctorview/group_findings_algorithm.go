package doctorview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
)

// GroupByProject files each finding under the first project it names: by the
// project's name in the check's own words, or by its directory anywhere in the
// check or its detail (a detail says "go is not on PATH" and means the tool).
// Findings that name no project (the toolchain, git, CI) form the workspace
// group, which comes first; the projects follow in the order given, and only
// those with findings appear. Within a group the findings keep their order, so
// problems stay on top.
func GroupByProject(findings []doctor.Finding, projects []ProjectRef) []Group {
	byProject := map[string][]int{}
	var workspace []int
	for i, f := range findings {
		if name := projectIn(f.Check, f.Detail, projects); name != "" {
			byProject[name] = append(byProject[name], i)
		} else {
			workspace = append(workspace, i)
		}
	}
	var groups []Group
	if len(workspace) > 0 {
		groups = append(groups, Group{Indexes: workspace})
	}
	for _, p := range projects {
		if idx := byProject[p.Name]; len(idx) > 0 {
			groups = append(groups, Group{Project: p.Name, Indexes: idx})
		}
	}
	return groups
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
