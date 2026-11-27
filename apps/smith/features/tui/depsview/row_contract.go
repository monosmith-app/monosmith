package depsview

// Row is a dependency with a newer release.
type Row struct {
	Name    string
	Plugin  string
	Current string
	Latest  string
	Level   string
	// Kinds say how projects declare the package: dev, prod, peer, optional,
	// build or managed.
	Kinds    []string
	Projects []string
	// URL is the package's page on its registry; empty when smith cannot know it.
	URL string
	// Declared is how each project declares the package (its spec, "^1.6.0"),
	// by project name, "" for the workspace root; nil when it is not known.
	Declared map[string]string
}

// PolicyLine is one result of the one-version policy.
type PolicyLine struct {
	Check   string
	OK      bool
	Warning bool
	Detail  string
	Remedy  string
}

// Result is what loading the view found: the shell receives it as a message.
type Result struct {
	Rows []Row
	// Policy is empty when smith.toml does not ask for the one-version policy.
	Policy []PolicyLine
	Err    error
}
