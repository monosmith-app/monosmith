// Package adoption is `smith adopt`: it reads a repository that is not yet an
// smith workspace (plain npm workspaces, Go modules, Python projects; an Nx
// workspace is smith convert's) and reports what bringing it under smith
// involves, each blocker and warning with the step that clears it. Two steps
// it can take itself: --tags creates the baseline release tags the lineage
// checks ask for, and --init writes smith.toml and the pipelines.
package adoption
