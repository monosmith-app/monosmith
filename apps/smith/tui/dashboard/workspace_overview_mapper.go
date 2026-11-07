package dashboard

import (
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/languagenames"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// maxLanguages is how many languages the workspace line names before "more".
const maxLanguages = 3

// workspaceText says which workspace this is: its name, its branch and what it
// holds ("monosmith · main · 34 projects (26 Go, 5 Python, 3 more)").
func workspaceText(f Facts) string {
	var parts []string
	for _, p := range []string{f.Workspace, f.Branch, projectsText(f)} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, " · ")
}

// projectsText says what the workspace holds: how many projects, and in which
// languages, most used first ("23 projects (18 TypeScript / JavaScript, 3 Python)").
func projectsText(f Facts) string {
	count := map[string]int{}
	var ids []string
	for _, l := range f.Languages {
		if l == "" {
			continue
		}
		if count[l] == 0 {
			ids = append(ids, l)
		}
		count[l]++
	}
	sort.SliceStable(ids, func(i, j int) bool { return count[ids[i]] > count[ids[j]] })
	var parts []string
	for i, id := range ids {
		if i == maxLanguages {
			parts = append(parts, words.Say("dashboard.projects.more", "count", len(ids)-maxLanguages))
			break
		}
		parts = append(parts, words.Say("dashboard.projects.language", "count", count[id], "language", languagenames.Of(id)))
	}
	text := words.Count("dashboard.project", f.Projects)
	if len(parts) == 0 {
		return text
	}
	return words.Say("dashboard.projects.summary", "projects", text, "languages", strings.Join(parts, ", "))
}
