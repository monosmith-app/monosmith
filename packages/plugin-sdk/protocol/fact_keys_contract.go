package protocol

// The well-known fact keys. A fact is a typed value under a namespaced key
// ("<namespace>.<name>"); the value of each key below has the type its
// comment names. A plugin may publish any other namespaced key, whose value
// can be any JSON value.
const (
	// FactModuleAliases maps an import alias to the project-relative
	// directory it stands for ({"@app": "src"}): an object of strings.
	FactModuleAliases = "module.aliases"
	// FactSourceRoots are the project-relative directories that hold source
	// code (["src"]): an array of strings.
	FactSourceRoots = "source.roots"
	// FactTestPatterns are the globs that select test files, relative to the
	// project (["src/**/*.spec.ts"]): an array of strings.
	FactTestPatterns = "test.patterns"
	// FactBuildOutputs are the project-relative directories a build writes
	// (["dist"]): an array of strings.
	FactBuildOutputs = "build.outputs"
	// FactRuntimeNode is the Node.js version range the project runs on
	// ("22"): a string.
	FactRuntimeNode = "runtime.node"
)
