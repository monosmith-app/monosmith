// Package workspace finds the workspace root and loads smith.toml: one small
// file whose every setting has a default, so most workspaces only name
// themselves. Overrides of plugin-inferred tasks live here too, and each one is
// recorded so `smith explain` can say where a value came from.
package workspace
