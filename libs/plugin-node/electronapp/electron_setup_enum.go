package electronapp

// Tool is the Electron setup of a project.
type Tool string

const (
	// None is a project that does not pack an Electron app.
	None Tool = ""
	// SmithForge is the app smith makes: its scripts build into the work
	// folder and run Electron Forge there (scripts/assemble.mjs).
	SmithForge Tool = "smith-forge"
	// Forge is an adopted Electron Forge setup (@electron-forge/cli).
	Forge Tool = "forge"
	// Builder is an adopted electron-builder setup.
	Builder Tool = "electron-builder"
)
