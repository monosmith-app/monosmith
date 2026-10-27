package formatting

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"

// Args returns the arguments rustfmt takes (after cargo fmt's `--`). When the
// crate, or a directory above it up to the workspace root, has a rustfmt.toml
// or .rustfmt.toml, rustfmt finds it itself and nothing is passed; otherwise
// smith's default is written to .smith/defaults/ and passed with --config-path.
// The second result says which of the two applies, for the task's description.
func Args(workspaceRoot, projectRoot string) (args []string, own bool, err error) {
	if tooldefaults.FindConfig(workspaceRoot, projectRoot, tooldefaults.HasFile("rustfmt.toml", ".rustfmt.toml")) {
		return nil, true, nil
	}
	rel, err := tooldefaults.WriteDefault(workspaceRoot, tooldefaults.Rustfmt.File, defaultConfig)
	if err != nil {
		return nil, false, err
	}
	return []string{"--config-path", "{workspaceRoot}/" + rel}, false, nil
}
