// Package taskrun runs a task graph: each task waits for its dependencies,
// replays from the cache when its key matches, and otherwise runs as a child
// process whose whole tree is killed on cancel. A failed task skips its
// dependents; independent tasks keep going unless --bail is set.
package taskrun
