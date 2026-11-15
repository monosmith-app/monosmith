package dependencies

import "strings"

// RootPlugin is the plugin that can add packages to the manifest at the
// workspace root (the package.json shared by every project). The other
// languages have no such plugin path yet, so a root target is refused for them.
const RootPlugin = "node"

// RootInstallCommand is the command that adds packages at the workspace root,
// for the plan and the confirmation.
//
// Parameters: packages are the names; dev is -D.
// Returns: the command text.
func RootInstallCommand(packages []string, dev bool) string {
	cmd := "npm install " + strings.Join(packages, " ")
	if dev {
		cmd += " --save-dev"
	}
	return cmd
}

// RootTarget is the -w value that names the workspace root: the package.json
// at the top of the workspace is not a project of the graph, so it has no name,
// and the directory "." is what stands for it. smith deps up lists its
// dependencies as "(root)".
const RootTarget = "."

// SplitRootTarget takes the workspace root out of a list of -w targets.
//
// Parameters: targets are the values of -w.
// Returns: whether the root was named, and the other targets in their order.
func SplitRootTarget(targets []string) (root bool, others []string) {
	for _, t := range targets {
		if t == RootTarget {
			root = true
			continue
		}
		others = append(others, t)
	}
	return root, others
}
