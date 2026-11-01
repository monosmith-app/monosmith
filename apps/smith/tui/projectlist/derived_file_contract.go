package projectlist

// The states of a derived file, as `smith derive` names them, plus
// StatusDiffers for a file that smith left alone because the project made it
// differ from its sibling projects.
const (
	StatusCurrent    = "current"
	StatusStale      = "stale"
	StatusMissing    = "missing"
	StatusUnrecorded = "unrecorded"
	StatusDrift      = "drift"
	StatusDiverged   = "diverged"
	StatusBroken     = "broken"
	StatusEjected    = "ejected"
	StatusDiffers    = "differs"
)

// DerivedFile is a file of a project that smith derives from the project graph
// or from the shape of the sibling projects, and the state it is in.
type DerivedFile struct {
	// Path is relative to the workspace root, with forward slashes.
	Path string
	// Renderer is the id of what writes it ("node.tsconfig.references").
	Renderer string
	// Status is one of the Status constants.
	Status string
	// Detail says in a sentence why the file is in that state.
	Detail string
	// Keys is true when smith owns only some keys of the file; only such a
	// file can be adopted with `smith derive --adopt`.
	Keys bool
}
