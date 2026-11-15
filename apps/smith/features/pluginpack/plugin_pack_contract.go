package pluginpack

// Plugin is what the packer needs of a manifest: the caller has already read
// plugin.json.
type Plugin struct {
	Name    string
	Version string
	// Command is the manifest's command, placeholders unexpanded.
	Command []string
}

// Result is a finished archive.
type Result struct {
	// Path is the archive, and DigestPath the file beside it holding its SHA-256.
	Path       string
	DigestPath string
	// SHA256 is the digest of the archive itself.
	SHA256 string
	// Files are the packed names in order, and Programs those the manifest's
	// command names (one per platform shipped).
	Files    []string
	Programs []string
	// Bytes is the size of the archive.
	Bytes int64
}
