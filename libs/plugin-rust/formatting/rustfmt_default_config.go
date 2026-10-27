package formatting

// defaultConfig is what rustfmt runs with when the repository configures none:
// rustfmt's own style, with the line endings pinned so a checkout formats the
// same on every OS. A repository's rustfmt.toml replaces it whole.
const defaultConfig = `# smith's default rustfmt configuration. Add a rustfmt.toml to replace it.
max_width = 100
newline_style = "Unix"
`

// DefaultConfig returns the text of smith's default configuration.
func DefaultConfig() string { return defaultConfig }
