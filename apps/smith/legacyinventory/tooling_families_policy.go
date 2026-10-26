package legacyinventory

import (
	"path"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
)

// Family is a kind of tooling a choice retires: all the files and packages of
// one tool that smith replaces.
type Family string

// The families. A choice retires a family when smith (or another tool) takes
// over its job: Lint when the lint config is smith's, Jest or Vitest when the
// other runner was chosen, Build when smith builds, Nx always at the full
// level.
const (
	FamilyLint   Family = "lint"
	FamilyJest   Family = "jest"
	FamilyVitest Family = "vitest"
	FamilyBuild  Family = "build"
	FamilyNx     Family = "nx"
)

// SmithLintPackage is the shared ESLint config smith uses; it is never part
// of the lint family that a lint choice retires.
const SmithLintPackage = "@monosmith/eslint-config"

// fileRule says which files belong to a family and what keeps one of them.
type fileRule struct {
	family Family
	match  func(rel string) bool
	// byName keeps the file while another remaining file mentions its name
	// (the way an eslint config imports a sibling config).
	byName bool
	// tool keeps the file while a remaining file mentions the tool that
	// loads it by convention (a rollup config while a script runs rollup).
	tool string
}

func baseOf(rel string) string { return path.Base(rel) }

func hasPrefixes(rel string, prefixes ...string) bool {
	b := baseOf(rel)
	for _, p := range prefixes {
		if strings.HasPrefix(b, p) {
			return true
		}
	}
	return false
}

var fileRules = []fileRule{
	{family: FamilyNx, byName: true, match: func(rel string) bool {
		b := baseOf(rel)
		return b == ".nxignore" || b == "mnci.json" ||
			(path.Dir(rel) == "." && strings.HasPrefix(b, ".mnci"))
	}},
	{family: FamilyLint, byName: true, match: func(rel string) bool {
		b := baseOf(rel)
		return hasPrefixes(rel, "eslint.config.", ".eslintrc") || b == ".eslintignore"
	}},
	{family: FamilyJest, byName: true, match: func(rel string) bool {
		return hasPrefixes(rel, "jest.config.", "jest.preset.", "jest.setup.", "jest.resolver.", "jest.asset-transform.")
	}},
	{family: FamilyVitest, byName: true, match: func(rel string) bool {
		return hasPrefixes(rel, "vitest.config.", "vitest.workspace.", "vitest.setup.")
	}},
	{family: FamilyBuild, tool: "rollup", match: func(rel string) bool {
		return hasPrefixes(rel, "rollup.config.", "rollup.base.")
	}},
	{family: FamilyBuild, tool: "webpack", match: func(rel string) bool {
		return hasPrefixes(rel, "webpack.config.", "webpack.base.")
	}},
	{family: FamilyBuild, tool: "swc", match: func(rel string) bool { return baseOf(rel) == ".swcrc" }},
	{family: FamilyBuild, tool: "babel", match: func(rel string) bool {
		return hasPrefixes(rel, ".babelrc", "babel.config.")
	}},
}

// fileFamily is the family a file belongs to, and its rule.
func fileFamily(rel string) (fileRule, bool) {
	if strings.HasPrefix(rel, "node_modules/") || strings.Contains(rel, "/node_modules/") {
		return fileRule{}, false
	}
	for _, r := range fileRules {
		if r.match(rel) {
			return r, true
		}
	}
	return fileRule{}, false
}

// packageFamily is the family a package belongs to. A package is in a family
// only by name: the list is what is known to be replaced, and a package that
// matches nothing is never offered for removal.
func packageFamily(name string) (Family, bool) {
	switch {
	case name == SmithLintPackage || name == "eslint":
		return "", false
	case name == "@mnci/eslint-config":
		return FamilyLint, true
	case convertplan.NxPackage(name) || strings.HasPrefix(name, "@mnci/"):
		return FamilyNx, true
	case strings.Contains(name, "eslint"):
		return FamilyLint, true
	case name == "jest" || strings.HasPrefix(name, "jest-") || strings.HasPrefix(name, "@jest/") ||
		name == "ts-jest" || name == "@swc/jest" || name == "babel-jest" || name == "@types/jest" || name == "@monosmith/jest-preset":
		return FamilyJest, true
	case name == "vitest" || strings.HasPrefix(name, "@vitest/") || strings.HasPrefix(name, "vitest-"):
		return FamilyVitest, true
	case buildPackage(name):
		return FamilyBuild, true
	}
	return "", false
}

func buildPackage(name string) bool {
	switch name {
	case "rollup", "webpack", "webpack-cli", "ts-loader", "swc-loader", "babel-loader", "@swc/cli", "tsup", "tsdx", "microbundle", "parcel":
		return true
	}
	return strings.HasPrefix(name, "@rollup/") || strings.HasPrefix(name, "rollup-plugin-")
}

// familyTokens are the words whose presence in a file that stays means the
// family's tool is still in use.
var familyTokens = map[Family][]string{
	FamilyJest:   {"jest"},
	FamilyVitest: {"vitest"},
}
