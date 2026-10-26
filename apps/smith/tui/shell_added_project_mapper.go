package tui

import (
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

// mainLanguage is the language most of the workspace's projects use: what a
// new project starts as; "" in a workspace with no projects.
func (s Shell) mainLanguage() string {
	var used []string
	for _, p := range s.list.Projects() {
		used = append(used, p.Language)
	}
	return languagenames.MostUsed(used)
}
