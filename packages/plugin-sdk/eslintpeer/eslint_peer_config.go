package eslintpeer

// ConfigPackage is smith's shared ESLint config, and ConfigRange the range
// that installs it.
const (
	ConfigPackage = "@monosmith/eslint-config"
	ConfigRange   = "^1.0.0"
)

// Range is the eslint ConfigPackage runs with: its peer dependency in
// packages/eslint-config/package.json, which a test compares.
const Range = "^10.8.1"

// Floor is the oldest eslint version Range accepts.
const Floor = "10.8.1"
