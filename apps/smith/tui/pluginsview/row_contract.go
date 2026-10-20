package pluginsview

// Row is one plugin as the Plugins view shows it.
type Row struct {
	Name    string
	Version string
	// Source is "built-in" or the directory an installed plugin was found in.
	Source  string
	Builtin bool
	// Switchable is true for a built-in plugin that smith.toml can turn on or off.
	Switchable bool
	Enabled    bool
	// Roles are the levels (language, deploy) and slots (lint) it fills.
	Roles    []string
	Requires []string
	// UsedBy are the projects that plugin discovered or deploys.
	UsedBy []string
}

// Result is what loading the view found: the shell receives it as a message.
type Result struct {
	Rows []Row
	Err  error
}
