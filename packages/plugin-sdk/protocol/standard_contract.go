package protocol

import "encoding/json"

// The ways a standard file is applied.
const (
	// StandardFile means the plugin gives the whole file: the host writes it
	// and keeps the file it replaces beside it as <path>.pre-smith.
	StandardFile = "file"
	// StandardKeys means the plugin gives the values of some top-level keys of
	// a JSON or TOML file (a manifest that also holds the dependencies, the
	// name and the version): the host sets those keys and keeps every other
	// key as it is. No copy is kept; the finding and the conversion summary say
	// what the keys held.
	StandardKeys = "keys"
)

// StandardParams asks a plugin for smith's standard version of the
// configuration files of its projects: the files `smith add` writes for a new
// project of the same kind, fitted to what the project already is. A plugin
// leaves out a file it cannot standardize without breaking the project (a
// browser project's tsconfig.json, a build file that holds the dependencies).
// The plugin writes nothing: the host compares the files with the ones on disk,
// asks, and writes (smith adopt, smith convert, smith doctor --fix).
type StandardParams struct {
	Projects []Project `json:"projects"`
}

// StandardResult carries the standard files. A file that already is as the
// plugin gives it is not a difference: the host compares.
type StandardResult struct {
	Files []StandardFileContent `json:"files"`
}

// StandardFileContent is smith's standard for one file.
type StandardFileContent struct {
	// Project is the project the file belongs to; empty for a file of the
	// workspace (the root tsconfig.base.json).
	Project string `json:"project,omitempty"`
	// Path is workspace-relative, with forward slashes.
	Path string `json:"path"`
	// Mode is StandardFile or StandardKeys.
	Mode string `json:"mode"`
	// Content is the whole file, for StandardFile.
	Content string `json:"content,omitempty"`
	// Format is "json" or "toml", for StandardKeys.
	Format string `json:"format,omitempty"`
	// Table is the TOML table that holds the keys ("tool.pytest.ini_options");
	// empty for the top level and for JSON.
	Table string `json:"table,omitempty"`
	// Keys are the top-level keys to set and their whole values as JSON, for
	// StandardKeys. A key's value replaces the one on disk, so a plugin that
	// adds to an object gives the object with its other members kept. The
	// members of an object value keep the plugin's order, which matters where
	// order is meaning (the conditions of a package.json "exports").
	Keys map[string]json.RawMessage `json:"keys,omitempty"`
	// Reason says in one sentence what the standard gives the team ("Ctrl+click
	// on an import opens the library's source, not its built .d.ts").
	Reason string `json:"reason"`
}
