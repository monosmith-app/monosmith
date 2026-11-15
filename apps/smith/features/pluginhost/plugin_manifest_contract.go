package pluginhost

// ManifestFile is the file that makes a directory an installed plugin.
const ManifestFile = "plugin.json"

// Manifest is plugin.json: how to run an installed plugin. Any install
// channel (a registry, a package manager, a copy) only has to put a directory
// holding one on the search path.
//
//	{
//	  "name": "kiwi",
//	  "version": "1.2.0",
//	  "description": "the kiwi language",
//	  "command": ["./bin/smith-plugin-kiwi-{os}-{arch}{exe}"],
//	  "env": ["NPM_TOKEN", "AWS_*"]
//	}
//
// A command path starting with ./ or ../ is relative to the plugin's
// directory; {os}, {arch} and {exe} (".exe" on Windows) are filled in. env
// lists the environment variables the plugin needs beyond the baseline every
// external plugin gets: names, or prefixes ending in *.
type Manifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Command     []string `json:"command"`
	Env         []string `json:"env,omitempty"`
}

// Installed is a plugin found on the search path.
type Installed struct {
	Manifest
	// Dir is the plugin's directory.
	Dir string `json:"dir"`
	// Argv is the command with its path resolved and placeholders filled.
	Argv []string `json:"argv"`
}
