package tooling

// Setup is what a toolchain plugin adds when a Node project is generated. Nothing
// already there is changed, except a dependency Raise names: every file,
// script, field and dependency is added only when absent.
type Setup struct {
	// RootFiles are written at the workspace root (path → content).
	RootFiles map[string]string
	// DevDependencies join the root package.json.
	DevDependencies map[string]string
	// Raise changes a root dependency the workspace already names (in
	// dependencies, devDependencies or optionalDependencies): the function
	// gets its range and returns the new range and true, or false to keep it.
	// It is for a tool the plugin's config cannot run with below some version.
	Raise map[string]func(current string) (string, bool)
	// RootFields join the top level of the root package.json when it has no
	// such field (npm "overrides" that settle a peer range, for example); a
	// field the workspace already has is its own and stays.
	RootFields map[string]any
	// Files are written in the project (project-relative path → content).
	Files map[string]string
	// Scripts join the project's package.json scripts.
	Scripts map[string]string
	// Fields join the top level of the project's package.json.
	Fields map[string]any
}
