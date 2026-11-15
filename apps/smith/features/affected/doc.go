// Package affected works out which projects a change can break: the projects
// owning changed files, every project that depends on them, and everything
// when a workspace-level file changed or the base cannot be resolved. Too wide
// costs minutes; too narrow reports green for code that was never checked.
package affected
