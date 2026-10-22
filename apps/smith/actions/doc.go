// Package actions is the one list of what smith can do: each action with the
// command it is, what it does to the workspace, whether it asks first, and
// which interface leaves it out and why.
//
// The command line, the interactive shell, the MCP server and the editor
// engine each present these actions their own way, but none of them keeps a
// list of its own: they read this one, and a test in each fails when an action
// is neither presented nor excused there. Presentation that belongs to one
// interface (a shortcut key, a menu position) stays with that interface and
// names the action it presents.
package actions
