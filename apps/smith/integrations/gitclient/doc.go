// Package gitclient is smith's only door to git. It shells out to the user's
// git, so hooks, config, worktrees and credentials behave exactly as they do
// at the terminal, and parses NUL-separated output so paths with spaces or
// unicode survive on every OS.
package gitclient
