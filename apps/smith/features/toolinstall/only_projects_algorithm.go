package toolinstall

import "slices"

// OnlyProjects keeps the tools that the named projects use, and the tools no
// project owns (git, the GitHub CLI, a tool named on the command line).
//
// Parameters: needs are the tools of the workspace; projects are the projects a
// command touches; nil keeps every tool.
// Returns: the tools the command can need.
func OnlyProjects(needs []Need, projects []string) []Need {
	if projects == nil {
		return needs
	}
	var out []Need
	for _, n := range needs {
		if len(n.Projects) == 0 || slices.ContainsFunc(n.Projects, func(p string) bool { return slices.Contains(projects, p) }) {
			out = append(out, n)
		}
	}
	return out
}
