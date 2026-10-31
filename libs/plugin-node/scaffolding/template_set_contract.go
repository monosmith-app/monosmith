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
	// Slots are lists of text that the templates read as
	// [[ range index .Slots "name" ]]. A framework or a style adds to a slot, so
	// the shared templates need no change for each.
	Slots map[string][]string
	// Scripts are the project's package.json scripts, read by a template as
	// .Scripts (sorted by name). They come from the tasks of the chosen choices.
	Scripts map[string]string
	// Dependencies are the project's package.json dependencies, read by a
	// template as .Dependencies (sorted by name).
	Dependencies map[string]string
}

// Pair is one name and value of a sorted map, as a template reads it. The
// value is escaped for a JSON string, so a template writes "[[.Value]]".
type Pair struct {
	Name, Value string
}
