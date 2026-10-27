package linting

// defaultConfig is what ruff runs with when the repository configures none.
// A repository's ruff.toml (or [tool.ruff] table) replaces it whole; it can
// start from this text, or point `extend` at another file.
const defaultConfig = `# smith's default ruff configuration. Add a ruff.toml to replace it.
line-length = 100
target-version = "py312"

[lint]
select = ["E", "F", "I", "UP", "B", "N"]
`

// DefaultConfig returns the text of smith's default configuration.
func DefaultConfig() string { return defaultConfig }
