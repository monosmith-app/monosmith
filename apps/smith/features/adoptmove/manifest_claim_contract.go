package adoptmove

// Claim is what one language plugin says about the projects it discovers: its
// language and the manifest file names it reads (exact names or path.Match
// patterns such as "*.csproj"). Detect identifies the project at the
// repository root from the claims of the plugins the session started, so
// adopt and the workspace know the same languages.
type Claim struct {
	// Language is the plugin's first language name, such as "node" or "php".
	Language string `json:"language"`
	// Manifests are the base names, or patterns of base names, the plugin
	// discovers projects by.
	Manifests []string `json:"manifests"`
}
