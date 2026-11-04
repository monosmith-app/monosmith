package tui

import "strings"

// uninstallDetails is the plan the confirmation of smith uninstall shows: the
// packages, and the manifest of each project they leave. The workspace root
// (-w .) is the package.json at the top of the workspace.
//
// Parameters: argv is the uninstall command line without the program name.
func uninstallDetails(argv []string) []string {
	var targets, packages []string
	for i := 1; i < len(argv); i++ {
		switch a := argv[i]; {
		case a == "-w" || a == "--workspace":
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
	lines := []string{"Removes " + strings.Join(packages, ", ") + " from the dependencies, or the development dependencies, of:"}
	for _, t := range targets {
		if t == "." {
			lines = append(lines, "  the workspace root (package.json at the top)")
			continue
		}
		lines = append(lines, "  "+t)
	}
	return append(lines, "The lock file changes with it. Nothing is committed unless you tick the box.")
}
