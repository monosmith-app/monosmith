package protocol

// ComponentsParams asks a plugin for the components one of its projects
// ships: what a CycloneDX SBOM lists.
type ComponentsParams struct {
	Project Project `json:"project"`
}

// Component is one dependency of a project.
type Component struct {
	Name string `json:"name"`
	// Version is the locked version, or the declared range when the project
	// has no lockfile.
	Version string `json:"version,omitempty"`
	// PURL is the package URL (https://github.com/package-url/purl-spec).
	PURL string `json:"purl"`
}

// ComponentsResult lists the components, read from a lockfile where the
// ecosystem has one and from the manifest otherwise. An empty list is an
// answer: the project declares no dependencies.
type ComponentsResult struct {
	Components []Component `json:"components"`
	// Source is the workspace-relative file the list was read from, "" when
	// the project has none.
	Source string `json:"source,omitempty"`
	// PURL is the project's own package URL when it has a registry identity
	// (without a version: the release supplies it); "" otherwise.
	PURL string `json:"purl,omitempty"`
}
