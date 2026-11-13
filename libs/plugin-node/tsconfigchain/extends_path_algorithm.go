package tsconfigchain

import "strings"

// Names of the root files of the chain.
const (
	// Base is the workspace's tsconfig.base.json: it extends the preset and holds
	// what every project shares.
	Base = "tsconfig.base.json"
	// Browser is the workspace's tsconfig.browser.json: the base, with the
	// options of code that a bundler builds.
	Browser = "tsconfig.browser.json"
	// Preset is the package the base extends.
	Preset = "@monosmith/tsconfig/node.json"
)

// Up leads from a project folder to the workspace root: "../../" for "apps/api".
// A folder at the root gives "".
func Up(dir string) string {
	dir = strings.Trim(strings.ReplaceAll(dir, "\\", "/"), "/")
	if dir == "" || dir == "." {
		return ""
	}
	return strings.Repeat("../", strings.Count(dir, "/")+1)
}

// Extends is the "extends" value that points a project at a root file of the
// chain, with the depth of the project: Extends("apps/api", Base) is
// "../../tsconfig.base.json".
func Extends(dir, file string) string {
	if Up(dir) == "" {
		return "./" + file
	}
	return Up(dir) + file
}
