package sbom

// Component is one dependency of a released project.
type Component struct {
	Name    string
	Version string
	// PURL is the package URL (pkg:golang/..., pkg:npm/...).
	PURL string
}

// Subject is the released project the SBOM describes.
type Subject struct {
	Name    string
	Version string
	// PURL is empty when the project has no registry identity.
	PURL string
}
