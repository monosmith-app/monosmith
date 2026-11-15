// Package taskgraph expands the tasks a user asked for into every task they
// depend on: "build" means this project's build first, "^build" means the
// build of every project this one imports first, reaching through projects
// that have no such task.
package taskgraph
