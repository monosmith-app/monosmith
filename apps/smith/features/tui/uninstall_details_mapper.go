package tui

import (
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"strings"
)

// uninstallDetails is the plan the confirmation of smith uninstall shows: the
// packages, and the manifest of each project they leave. The workspace root
// (-w .) is the package.json at the top of the workspace.
//
// Parameters: argv is the uninstall command line without the program name.
func uninstallDetails(argv []string) []string {
	var targets, packages []string
	for i := 1; i < len(argv); i++ {
		switch a := argv[i]; {
		case flagvocabulary.Is("project", a):
			if i+1 < len(argv) {
				i++
				targets = append(targets, argv[i])
			}
		case strings.HasPrefix(a, "-"):
		default:
			packages = append(packages, a)
		}
	}
	if len(packages) == 0 || len(targets) == 0 {
		return nil
	}
	lines := []string{words.Say("tui.uninstall_details.removes_from_the_dependencies_or_the", "list", strings.Join(packages, ", "))}
	for _, t := range targets {
		if t == "." {
			lines = append(lines, words.Say("tui.uninstall_details.the_workspace_root_package_json_at"))
			continue
		}
		lines = append(lines, "  "+t)
	}
	return append(lines, words.Say("tui.uninstall_details.the_lock_file_changes_with_it"))
}
