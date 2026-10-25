package pluginadmit

import "github.com/monosmith-app/monosmith/apps/smith/plugintrust"

// Plugin is an external plugin the workspace needs, resolved to what would
// run.
type Plugin struct {
	Name string `json:"name"`
	// Command is the argv that starts it.
	Command []string `json:"command"`
	// Source says where the plugin comes from: "smith.toml",
	// ".smith/plugins", "SMITH_PLUGIN_PATH" or "$SMITH_HOME/plugins".
	Source string `json:"source"`
	// Where is the file or directory behind Source, when there is one.
	Where string `json:"where,omitempty"`
	// ManifestEnv are the variables the plugin's own plugin.json declares.
	ManifestEnv []string `json:"manifestEnv,omitempty"`
	// WorkspaceEnv are the variables smith.toml grants it.
	WorkspaceEnv []string `json:"workspaceEnv,omitempty"`
	// Pin is the sha256 smith.toml gives it, if any.
	Pin string `json:"pin,omitempty"`
}

// Status is what `smith plugin list` and `smith doctor` show for an external
// plugin. Env lists names and patterns only, never values.
type Status struct {
	Plugin
	// Env are the variables it may see beyond the baseline.
	Env []string `json:"env"`
	// Trusted is whether it may start now: on a machine, the user's recorded
	// decision still matches; in CI, the pin matches.
	Trusted bool `json:"trusted"`
	// Pinned is whether smith.toml pins its sha256.
	Pinned bool `json:"pinned"`
	// Reason and Detail say why an untrusted plugin is not trusted.
	Reason plugintrust.Reason `json:"reason,omitempty"`
	Detail string             `json:"detail,omitempty"`
	// Fingerprint is its command and files as they are now; empty when the
	// program was not found.
	Fingerprint plugintrust.Fingerprint `json:"fingerprint"`
	// Problem is set when the command could not be measured.
	Problem string `json:"problem,omitempty"`
}
