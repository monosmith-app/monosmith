package scaffolding

// TemplateSet is what a Node project starts from: the files of every
// project, those of each kind, root files written once, and the root
// devDependencies it needs. The node plugin's is plain JavaScript; the
// typescript plugin brings its own. Templates use [[ ]] delimiters with
// .Dir, .Name, .Package and .Stack.
type TemplateSet struct {
	Slice   map[string]string
	App     map[string]string
	Lib     map[string]string
	Package map[string]string
	// Root files are written at the workspace root when absent.
	Root map[string]string
	// DevDependencies join the root package.json when absent.
	DevDependencies map[string]string
	// Overrides join the root package.json "overrides" (npm) when absent: a
	// version npm must use for a dependency a stack's own packages reach
	// through an optional peer, so adding the stack to a workspace that
	// already has other tooling can still install.
	Overrides map[string]string
}
