// Package session opens a workspace for a command: it finds the root, loads
// smith.toml, starts the plugins and builds the project graph, and closes the
// plugins when the command is done.
package session
