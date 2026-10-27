package linting

// defaultConfig is what PHPStan runs with when the package configures none.
// The paths to analyse are passed on the command line, so the file holds only
// the rules. A package's own phpstan.neon replaces it whole.
const defaultConfig = `# smith's default PHPStan configuration. Add a phpstan.neon to replace it.
parameters:
    level: 8
`

// DefaultConfig returns the text of smith's default configuration.
func DefaultConfig() string { return defaultConfig }
