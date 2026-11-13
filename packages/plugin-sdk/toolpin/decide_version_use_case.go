package toolpin

// Version is the toolchain version a workspace uses for a label (node, go,
// python, rustc, java, dotnet, ruby, php): the [toolchains] pin in smith.toml,
// else the language's native pin file, else any extra candidate the caller
// found (for example the newest go.mod directive), else smith's Default. It is
// empty only for a label with no pin and no default.
//
// workspaceRoot is the folder holding smith.toml. extra are candidates the
// plugin read itself, tried in order after the native file.
func Version(workspaceRoot, label string, extra ...string) string {
	if v := smithPin(workspaceRoot, label); v != "" {
		return v
	}
	if v := nativePin(workspaceRoot, label); v != "" {
		return v
	}
	for _, v := range extra {
		if v != "" {
			return v
		}
	}
	return Default(label)
}
