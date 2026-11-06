package dashboard

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/languagenames"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
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
			parts = append(parts, words.Say("dashboard.projects.more", "count", len(ids)-maxLanguages))
			break
		}
		parts = append(parts, words.Say("dashboard.projects.language", "count", count[id], "language", languagenames.Of(id)))
	}
	text := words.Count("dashboard.project", f.Projects)
	if len(parts) == 0 {
		return text
	}
	return text + ": " + strings.Join(parts, " · ")
}

// whereToGo is the list of places the tabs lead to, one line each, so a
// first-time user sees what the rest of smith is for. It starts with the two
// ways to make a project, as buttons; at maps the line of each button (an
// index into the lines) to its command.
func whereToGo(th theme.Theme, w int) (lines []string, at map[int]string) {
	places := []struct{ id, what string }{
		{"view.projects", words.Say("dashboard.place.projects")},
		{"view.run", words.Say("dashboard.place.run")},
		{"view.doctor", words.Say("dashboard.place.doctor")},
		{"view.deps", words.Say("dashboard.place.deps")},
		{"view.release", words.Say("dashboard.place.release")},
	}
	lines, at = []string{"", " " + th.Muted(words.Say("dashboard.where"))}, map[int]string{}
	for _, b := range []struct{ id, label, what string }{
		{"file.add", words.Say("dashboard.add.label"), words.Say("dashboard.add.what")},
		{"file.compose", words.Say("dashboard.compose.label"), words.Say("dashboard.compose.what")},
	} {
		at[len(lines)] = b.id
		button := button.Render(th, b.label, commands.ShortcutOf(b.id), button.Rest)
		lines = append(lines, "  "+button+" "+th.Muted(b.what))
	}
	for _, p := range places {
		c, ok := commands.ByID(p.id)
		if !ok {
			continue
		}
		lines = append(lines, pane.Wrap("  "+th.Accent(pane.Cell(c.Shortcut, 3))+pane.Cell(c.Title, 14)+th.Muted(p.what), w, 19)...)
	}
	return lines, at
}
