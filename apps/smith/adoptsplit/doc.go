// Package adoptsplit says, for each language, what of a single-project
// repository belongs to the workspace and what belongs to the project, and
// cuts the files that hold both (ADR 0090). Each language declares Rules: the
// top-level files that stay at the workspace root (tool configuration,
// lockfiles), a Sort that reads the manifest before the move and names the
// tool dependencies, and a Split that edits the moved files afterwards. A
// language without a Sort or Split moves its manifest whole. The package
// reads and writes files and knows nothing of git or of questions: adoptmove
// plans the move and asks, adoption runs it.
package adoptsplit
