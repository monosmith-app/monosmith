package depsview

// Row is a dependency with a newer release.
type Row struct {
	Name     string
	Plugin   string
	Current  string
	Latest   string
	Level    string
	// Kinds say how projects declare the package: dev, prod, peer, optional,
	// build or managed.
	Kinds    []string
	Projects []string
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
