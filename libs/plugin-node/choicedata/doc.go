// Package choicedata reads the choices a Node plugin ships as data (ADR 0045):
// one TOML file per choice, a version catalog, and an optional base and list of
// published stacks, all embedded in the plugin. A choice names an axis,
// requires and conflicts, dependencies by name, tasks, template slots and
// files. The package checks the files when it loads them, publishes the
// choices over the protocol, and merges the chosen ones into a Plan that the
// plugin hands to the scaffolding.
package choicedata
