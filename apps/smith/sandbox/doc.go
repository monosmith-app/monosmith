// Package sandbox runs a mutating smith command in a temporary copy of the
// work tree (a detached git worktree plus the uncommitted changes), runs
// checks on the result, and offers to run the same command for real
// (`smith try`, ADR 0057). The real tree is not touched before the person
// confirms.
package sandbox
