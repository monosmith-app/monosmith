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
	// Sources are the patterns of source file names the plugin recognises in
	// a folder without its manifest (*.ts, *.py), for the manifest offer.
	Sources []string `json:"sources,omitempty"`
	// Ignores are the folder-relative gitignore patterns of what the
	// language's tools write into a project folder (bin/, node_modules/).
	Ignores []string `json:"ignores,omitempty"`
}
