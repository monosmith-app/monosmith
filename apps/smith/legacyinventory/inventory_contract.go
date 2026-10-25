package legacyinventory

// LegacyFile is a file of a retired tool that nothing remaining needs.
type LegacyFile struct {
	Path   string `json:"path"`
	Family Family `json:"family"`
}

// LegacyPackage is a package of a retired tool that nothing remaining
// imports or runs.
type LegacyPackage struct {
	Name   string `json:"name"`
	Family Family `json:"family"`
	// In are the package.json files that declare it.
	In []string `json:"in"`
}

// Kept is a file or package that looks like the old tooling but stays, and
// why: the user is told, and can decide.
type Kept struct {
	What   string `json:"what"`
	Family Family `json:"family"`
	Why    string `json:"why"`
}

// Inventory is what a scan found.
type Inventory struct {
	Files    []LegacyFile    `json:"files,omitempty"`
	Packages []LegacyPackage `json:"packages,omitempty"`
	Kept     []Kept          `json:"kept,omitempty"`
}

// Usage is the tooling a workspace uses today; the choices default to it.
type Usage struct {
	// ESLint is true when the workspace lints with ESLint.
	ESLint bool `json:"eslint"`
	// LintConfig is the root lint config file, "" when there is none.
	LintConfig string `json:"lintConfig,omitempty"`
	// SharedLintConfig is the shared config package the root lint config
	// imports (@mnci/eslint-config), "" for a config written by hand.
	SharedLintConfig string `json:"sharedLintConfig,omitempty"`
	Jest             bool   `json:"jest"`
	Vitest           bool   `json:"vitest"`
}
