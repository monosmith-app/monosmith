package linting

// defaultConfig is what golangci-lint runs with when the repository has no
// .golangci.{yml,yaml,toml,json} of its own: the standard linters, minus the
// noise of writing to a terminal. A repository's own file replaces it whole
// (golangci-lint has no "extends"), so copy this to start from it.
const defaultConfig = `# smith's default golangci-lint configuration. Add a .golangci.yml to replace it.
version: "2"
linters:
  settings:
    errcheck:
      # Writing to the terminal or a cobra command's output cannot usefully fail.
      exclude-functions:
        - fmt.Fprint
        - fmt.Fprintf
        - fmt.Fprintln
`

// DefaultConfig returns the text of smith's default configuration.
func DefaultConfig() string { return defaultConfig }
