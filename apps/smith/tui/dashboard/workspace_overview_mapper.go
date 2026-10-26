package dashboard

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/languagenames"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// maxLanguages is how many languages the Projects row names before "and more".
const maxLanguages = 4

// projectsText says what the workspace holds: how many projects, and in which
// languages, most used first ("23 projects: 18 TypeScript / JavaScript · 3 Python").
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
			parts = append(parts, fmt.Sprintf("%d more", len(ids)-maxLanguages))
			break
		}
		parts = append(parts, fmt.Sprintf("%d %s", count[id], languagenames.Of(id)))
	}
	text := plural(f.Projects, "project", "projects")
	if len(parts) == 0 {
		return text
	}
	return text + ": " + strings.Join(parts, " · ")
}

func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}

// whereToGo is the list of places the tabs lead to, one line each, so a
// first-time user sees what the rest of smith is for.
func whereToGo(th theme.Theme, w int) []string {
	places := []struct{ id, what string }{
		{"view.projects", "every project, what it needs and what depends on it"},
		{"view.run", "build, test and lint, and watch them run"},
		{"view.doctor", "problems in the setup, with a fix for most"},
		{"view.deps", "newer versions of what the projects install"},
		{"view.release", "what the next release would publish"},
	}
	lines := []string{"", " " + th.Muted("Where to go")}
	for _, p := range places {
		c, ok := commands.ByID(p.id)
		if !ok {
			continue
		}
		lines = append(lines, pane.Wrap("  "+th.Accent(pane.Cell(c.Shortcut, 3))+pane.Cell(c.Title, 14)+th.Muted(p.what), w, 19)...)
	}
	return lines
}
