// Package projectgraph builds the workspace's project graph: projects come
// from plugins, edges come from matching what each project imports to what
// another exports, and smith.toml overrides are applied on top with their
// source recorded.
package projectgraph
