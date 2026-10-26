package imports

// Ref is one import of the project's own code: the path it names, split
// into folders and a last name, with the folder the project's sources are
// rooted at taken off where the language spells it that way.
type Ref struct {
	// Spec is the import as written, for messages.
	Spec string
	// Segments is the path it names ("domain", "order").
	Segments []string
}

// Project is what a language's imports are read against: facts about the
// project that tell its own code from a library's.
type Project struct {
	// GoModule is the module path in go.mod.
	GoModule string
	// DartPackage is the package name in pubspec.yaml.
	DartPackage string
	// PythonRoots are the top-level packages and modules of the project.
	PythonRoots map[string]bool
}

// Language is how one language's sources are found and read.
type Language struct {
	// Name is the language as its plugin names it.
	Name string
	// Extensions are the file extensions of its sources, with the dot.
	Extensions []string
	// Inputs are the globs the check reads, which key its cache.
	Inputs []string
	// IsTest reports whether a file, by its name, is a test: tests may
	// cross layers on purpose, so they are not checked.
	IsTest func(name string) bool
	// Extract lists the imports of the project's own code in one file; rel
	// is the file's path from the project root, forward slashes.
	Extract func(p Project, rel, text string) []Ref
}
