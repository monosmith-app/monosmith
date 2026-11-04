package dependencies

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
