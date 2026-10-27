package typechecking

// defaultConfig is what mypy runs with when the repository configures none.
// It checks untyped code too but asks for no annotations, so a project can
// adopt types gradually. A repository's own configuration replaces it whole.
const defaultConfig = `# smith's default mypy configuration. Add a mypy.ini to replace it.
[mypy]
python_version = 3.12
check_untyped_defs = True
ignore_missing_imports = True
explicit_package_bases = True
warn_redundant_casts = True
warn_unused_ignores = True
`

// DefaultConfig returns the text of smith's default configuration.
func DefaultConfig() string { return defaultConfig }
