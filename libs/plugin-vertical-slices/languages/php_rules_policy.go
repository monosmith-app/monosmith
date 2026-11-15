package languages

import (
	"path"
	"regexp"
	"strings"
)

var (
	phpTest = regexp.MustCompile(`^` + pascalName + `Test$`)
	// phpEntry are the classes a src/ root may hold (a framework kernel).
	phpEntry = map[string]bool{"Kernel": true, "Application": true, "App": true}
	// phpTestSetup are the files PHPUnit and Pest load before the tests.
	phpTestSetup = map[string]bool{"bootstrap.php": true, "Pest.php": true, "TestCase.php": true}
)

// phpRules follow PSR-1 and PSR-4: a file holds one class named like it,
// in PascalCase, and folders are PascalCase namespaces. src/ holds the
// slices, each a folder of <Name><Role>.php classes (PlaceOrderUseCase.php);
// tests mirror src/ under tests/ as <Name>Test.php, with test doubles in a
// Testing/ folder.
var phpRules = Rules{
	Name:   "php",
	Roots:  []string{"src", ""},
	Inputs: []string{"src/**/*.php", "tests/**/*.php"},
	Skip:   func(name string, in Dir) bool { return in.Rel == "" && name != "src" && name != "tests" },
	Folder: func(name string, in Dir) string {
		// src/ and tests/ are named by Composer's autoload, not namespaces.
		if in.Rel == "" || pascalCase.MatchString(name) {
			return ""
		}
		return "a PHP folder is a PascalCase namespace (PSR-4)"
	},
	Source: func(name string, in Dir) bool { return in.Rel != "" && strings.HasSuffix(name, ".php") },
	File:   PhpFile,
	Production: func(name string, in Dir) bool {
		return strings.HasPrefix(in.Rel, "src/") && !pascalBarrel(name, in)
	},
}

// PhpFile says why a PHP file breaks the rules, or "" when it does not.
func PhpFile(name string, in Dir) string {
	stem := strings.TrimSuffix(name, path.Ext(name))
	switch {
	case in.Rel == "tests" || strings.HasPrefix(in.Rel, "tests/"):
		if phpTest.MatchString(stem) || phpTestSetup[name] || (pascalTesting(in) && pascalCase.MatchString(stem)) {
			return ""
		}
		return "a test is <Name>Test.php, mirroring src/; test doubles go in a Testing/ folder"
	case in.Rel == "src":
		if phpEntry[stem] {
			return ""
		}
		return "src/ holds the slice folders; move this into the slice whose outcome it serves"
	case pascalBarrel(name, in) || pascalRole.MatchString(stem):
		return ""
	}
	return "a class in a slice is <Name><Role>.php (roles: " + strings.Join(pascalRoles(), ", ") + ")"
}
