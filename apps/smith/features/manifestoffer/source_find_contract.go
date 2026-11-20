package manifestoffer

// None is the --manifest answer that writes no manifest.
const None = "none"

// Find is what one language plugin recognises in a folder that has none of
// its manifests.
type Find struct {
	// Language is the plugin's first language name ("node", "python", "dotnet").
	Language string
	// Manifest is the manifest the plugin knows a project by, as a person
	// reads it: "package.json", or "a .csproj" for a pattern.
	Manifest string
	// Files are the source files found, relative to the folder, with forward
	// slashes, in order.
	Files []string
}

// Nested is a project whose manifest is below the top of a source.
type Nested struct {
	// Path is the project's folder, relative to the source, with forward slashes.
	Path string
	// Manifest is the manifest's file name.
	Manifest string
	// Language is the plugin's first language name.
	Language string
}
