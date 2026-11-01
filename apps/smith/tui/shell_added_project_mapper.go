package tui

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/languagenames"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// addedProject is the name of the project a `smith add <kind> <name>` line
// makes.
func addedProject(argv []string) (string, bool) {
	if len(argv) < 3 || argv[0] != "add" || strings.HasPrefix(argv[2], "-") {
		return "", false
	}
	return argv[2], true
}

// addedNextSteps is the line the Run tab shows under a finished add: what to
// do with the project just made, with the keys that do it.
func addedNextSteps(th theme.Theme, name string) string {
	return " " + th.Muted("Next: ") +
		th.Accent(commands.ShortcutOf("run.test")) + th.Muted(" runs its tests · ") +
		th.Accent(commands.ShortcutOf("file.add")) + th.Muted(" adds another · ") +
		th.Accent(commands.ShortcutOf("projects.install-package")) + th.Muted(" installs a package into "+name)
}

// mainLanguageWhy says in one line why the Add form starts on the language it
// does: how many of the workspace's projects use it; "" with no projects.
func (s Shell) mainLanguageWhy() string {
	main := s.mainLanguage()
	if main == "" {
		return ""
	}
	n, total := 0, 0
	for _, p := range s.list.Projects() {
		total++
		if p.Language == main {
			n++
		}
	}
	return fmt.Sprintf("Preselected: %d of the %d projects here use %s.", n, total, languagenames.Of(main))
}

// mainLanguage is the language most of the workspace's projects use: what a
// new project starts as; "" in a workspace with no projects.
func (s Shell) mainLanguage() string {
	var used []string
	for _, p := range s.list.Projects() {
		used = append(used, p.Language)
	}
	return languagenames.MostUsed(used)
}
