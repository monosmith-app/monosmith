package tooling

// Setup is what a toolchain plugin adds when a Node project is generated. Nothing
// already there is changed: every file, script, field and dependency is
// added only when absent.
type Setup struct {
	// RootFiles are written at the workspace root (path → content).
	RootFiles map[string]string
	// DevDependencies join the root package.json.
	DevDependencies map[string]string
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
