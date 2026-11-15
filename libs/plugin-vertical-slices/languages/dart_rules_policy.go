package languages

import (
	"path"
	"regexp"
	"strings"
)

var (
	dartRoleFile  = regexp.MustCompile(`^` + snakeName + `_(?:` + roleAlternation + `)\.dart$`)
	dartTestFile  = regexp.MustCompile(`^` + snakeName + `_test\.dart$`)
	dartSnakeFile = regexp.MustCompile(`^` + snakeName + `\.dart$`)
	// dartGenerated are code generators' outputs (build_runner, freezed,
	// mockito, auto_route), which nobody names by hand.
	dartGenerated = regexp.MustCompile(`\.(?:g|freezed|mocks|gr|config)\.dart$`)
)

// dartFolders are the package folders the rules check; the rest (platform
// folders, example/, tool/, build output) hold no slices.
var dartFolders = map[string]bool{"lib": true, "bin": true, "test": true, "integration_test": true}

// dartRules follow Effective Dart: files and folders are
// lowercase_with_underscores. lib/ holds the package's public libraries, and
// the slices under lib/src/: each file is <name>_<role>.dart, and a slice's
// public API is the barrel named after its folder (lib/src/orders/orders.dart).
// Tests mirror lib/ under test/ as <basename>_test.dart, with test doubles
// in a testing/ folder; bin/ holds entry points.
var dartRules = Rules{
	Name:   "dart",
	Roots:  []string{"lib"},
	Inputs: []string{"lib/**/*.dart", "bin/**/*.dart", "test/**/*.dart", "integration_test/**/*.dart"},
	Skip:   func(name string, in Dir) bool { return in.Rel == "" && !dartFolders[name] },
	Folder: func(name string, _ Dir) string {
		if !snakeCase.MatchString(name) {
			return "a Dart folder is lowercase_with_underscores"
		}
		return ""
	},
	Source: func(name string, in Dir) bool {
		return in.Rel != "" && strings.HasSuffix(name, ".dart") && !dartGenerated.MatchString(name)
	},
	File:       DartFile,
	Production: func(name string, in Dir) bool { return dartArea(in) == "lib" && !dartBarrel(name, in) },
}

// DartFile says why a Dart file breaks the rules, or "" when it does not.
func DartFile(name string, in Dir) string {
	switch area := dartArea(in); {
	case area == "test" || area == "integration_test":
		if dartTestFile.MatchString(name) || (inTesting(in) && dartSnakeFile.MatchString(name)) {
			return ""
		}
		return "a test is <basename>_test.dart, mirroring lib/; test doubles go in a testing/ folder"
	case area == "bin" || in.Rel == "lib":
		if dartSnakeFile.MatchString(name) {
			return ""
		}
		return "a Dart file is lowercase_with_underscores.dart"
	case in.Rel == "lib/src":
		return "lib/src holds slice folders; move this into the slice whose outcome it serves"
	case dartBarrel(name, in) || dartRoleFile.MatchString(name):
		return ""
	}
	return "a Dart file in a slice is <name>_<role>.dart (roles: " + strings.Join(Roles, ", ") + "), or the slice's barrel " + path.Base(in.Rel) + ".dart"
}

// dartArea is the package folder a folder is in: lib, bin, test or
// integration_test.
func dartArea(in Dir) string {
	area, _, _ := strings.Cut(in.Rel, "/")
	return area
}

// dartBarrel reports whether a file is its slice's public API: named after
// the folder it is in.
func dartBarrel(name string, in Dir) bool {
	return in.Rel != "lib" && name == path.Base(in.Rel)+".dart"
}
