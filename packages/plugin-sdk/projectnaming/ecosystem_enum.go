package projectnaming

// Ecosystem is the naming syntax a project name is built for.
type Ecosystem string

// The ecosystems of the language plugins. The first group puts the workspace
// namespace before the project name. The second group names an identifier in
// code (a module or class), where a namespace prefix would only lengthen what
// the person types in an import, so they use the project name alone.
const (
	Npm      Ecosystem = "npm"      // @acme/core
	Cargo    Ecosystem = "cargo"    // acme-core
	Python   Ecosystem = "python"   // acme-core
	Maven    Ecosystem = "maven"    // artifactId acme-core
	Dart     Ecosystem = "dart"     // acme_core
	CMake    Ecosystem = "cmake"    // acme_core
	Gem      Ecosystem = "gem"      // acme_core
	Composer Ecosystem = "composer" // acme/core
	DotNet   Ecosystem = "dotnet"   // Acme.Core

	Swift Ecosystem = "swift" // module Core
	ObjC  Ecosystem = "objc"  // pod and class Core
	VBA   Ecosystem = "vba"   // module core
	Lua   Ecosystem = "lua"   // rock core
)
