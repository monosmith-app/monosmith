// Package operation makes a mutating smith command one commit and lets
// `smith undo` revert it. With [workspace] commit_operations = true, a command
// that starts in a git work tree with no uncommitted change commits what it
// wrote as one conventional commit and records it in .smith/operations.json.
// smith undo reverts the last recorded operation with git revert (a new
// commit; history is never rewritten) and refuses when a later commit or an
// uncommitted change touches the same files. See ADR 0047.
package operation
