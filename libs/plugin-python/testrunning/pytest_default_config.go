package testrunning

// DefaultFile is the name smith gives its default under .smith/defaults/.
const DefaultFile = "pytest.ini"

// defaultConfig is what pytest runs with when the repository configures none.
// A repository's own configuration replaces it whole.
const defaultConfig = `# smith's default pytest configuration. Add a pytest.ini to replace it.
[pytest]
addopts = -ra --strict-markers
xfail_strict = true
testpaths = tests
`

// DefaultConfig returns the text of smith's default configuration.
func DefaultConfig() string { return defaultConfig }
