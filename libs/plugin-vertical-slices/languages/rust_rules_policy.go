package languages

import (
	"regexp"
	"strings"
)

var (
	rustRoleFile  = regexp.MustCompile(`^` + snakeName + `_(?:` + roleAlternation + `)\.rs$`)
	rustTestFile  = regexp.MustCompile(`^` + snakeName + `_test\.rs$`)
	rustSnakeFile = regexp.MustCompile(`^` + snakeName + `\.rs$`)
)

// rustRules follow the Rust API guidelines: modules (files and folders) are
// snake_case. src/ holds main.rs and lib.rs; a slice is a module folder
// under src/ whose public API is its mod.rs (or the <slice>.rs file beside
// the folder), holding <name>_<role>.rs files. Binaries in src/bin/ are
// entry points. Integration tests in tests/ are <name>_test.rs, with test
// doubles in a testing/ folder; unit tests stay inside the file they test.
var rustRules = Rules{
	Name:   "rust",
	Roots:  []string{"src"},
	Inputs: []string{"src/**/*.rs", "tests/**/*.rs"},
	Skip:   func(name string, in Dir) bool { return in.Rel == "" && name != "src" && name != "tests" },
	Folder: func(name string, _ Dir) string {
		if !snakeCase.MatchString(name) {
			return "a Rust module folder is snake_case"
		}
		return ""
	},
	Source: func(name string, in Dir) bool { return in.Rel != "" && strings.HasSuffix(name, ".rs") },
	File:   RustFile,
	Production: func(name string, in Dir) bool {
		return strings.HasPrefix(in.Rel, "src") && !rustModuleRoot(name, in)
	},
}

// RustFile says why a Rust file breaks the rules, or "" when it does not.
func RustFile(name string, in Dir) string {
	switch {
	case in.Rel == "tests" || strings.HasPrefix(in.Rel, "tests/"):
		if rustTestFile.MatchString(name) || name == "mod.rs" || (inTesting(in) && rustSnakeFile.MatchString(name)) {
			return ""
		}
		return "an integration test is <name>_test.rs; test doubles go in a testing/ folder"
	case in.Rel == "src/bin" || strings.HasPrefix(in.Rel, "src/bin/"):
		if rustSnakeFile.MatchString(name) {
			return ""
		}
		return "a binary in src/bin is snake_case.rs"
	case rustModuleRoot(name, in):
		return ""
	case in.Rel == "src":
		return "src/ holds main.rs, lib.rs and slice modules; move this into the slice whose outcome it serves"
	case rustRoleFile.MatchString(name):
		return ""
	}
	return "a Rust file in a slice is <name>_<role>.rs (roles: " + strings.Join(Roles, ", ") + "), or its mod.rs"
}

// rustModuleRoot reports whether a file is a crate root or a module's
// public API: main.rs and lib.rs in src/, a mod.rs, or <module>.rs beside
// its <module>/ folder.
func rustModuleRoot(name string, in Dir) bool {
	switch {
	case name == "mod.rs":
		return true
	case in.Rel == "src" && (name == "main.rs" || name == "lib.rs"):
		return true
	}
	return in.Folders[strings.TrimSuffix(name, ".rs")]
}
