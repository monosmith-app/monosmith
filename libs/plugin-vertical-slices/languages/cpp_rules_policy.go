package languages

import (
	"path"
	"regexp"
	"strings"
)

// cppExtensions are C and C++ sources and headers.
var cppExtensions = map[string]bool{".cpp": true, ".cc": true, ".cxx": true, ".c": true, ".h": true, ".hpp": true, ".hh": true, ".hxx": true, ".ipp": true}

var (
	cppSnakeRole  = regexp.MustCompile(`^` + snakeName + `_(?:` + roleAlternation + `)$`)
	cppSnakeTest  = regexp.MustCompile(`^` + snakeName + `_tests?$`)
	cppPascalTest = regexp.MustCompile(`^` + pascalName + `Tests?$`)
	cppPascal     = regexp.MustCompile(`^` + pascalName + `$`)
)

// cppRules accept both common spellings, Google style (snake_case files) and
// LLVM or Qt style (PascalCase), as long as a folder keeps to one of them.
// src/ holds the entry point and the slices, each a folder of <name>_<role>
// (or <Name><Role>) files; include/<target>/ holds the public headers, a
// slice's API being the header named after its folder; tests in tests/ are
// <name>_test.cpp (or <Name>Test.cpp), test doubles in a testing/ folder.
var cppRules = Rules{
	Name:   "cpp",
	Roots:  []string{"src", ""},
	Inputs: []string{"src/**", "include/**", "tests/**", "test/**"},
	Skip: func(name string, in Dir) bool {
		return in.Rel == "" && name != "src" && name != "include" && name != "tests" && name != "test"
	},
	Folder: func(name string, _ Dir) string {
		if !snakeCase.MatchString(name) && !cppPascal.MatchString(name) {
			return "a C++ folder is snake_case or PascalCase"
		}
		return ""
	},
	Source: func(name string, in Dir) bool { return in.Rel != "" && cppExtensions[path.Ext(name)] },
	File:   CppFile,
	Production: func(name string, in Dir) bool {
		area, _, _ := strings.Cut(in.Rel, "/")
		return area == "src" && in.Rel != "src" && !cppBarrel(name, in)
	},
}

// CppFile says why a C or C++ file breaks the rules, or "" when it does not.
func CppFile(name string, in Dir) string {
	stem := strings.TrimSuffix(name, path.Ext(name))
	if why := cppMixed(stem, in); why != "" {
		return why
	}
	area, _, _ := strings.Cut(in.Rel, "/")
	switch {
	case area == "tests" || area == "test":
		if cppSnakeTest.MatchString(stem) || cppPascalTest.MatchString(stem) || inTesting(in) {
			return ""
		}
		return "a test is <name>_test.cpp or <Name>Test.cpp; test doubles go in a testing/ folder"
	case in.Rel == "src":
		if strings.EqualFold(stem, "main") {
			return ""
		}
		return "src/ holds main.cpp and slice folders; move this into the slice whose outcome it serves"
	case area == "include" && strings.Count(in.Rel, "/") <= 1:
		if cppBarrel(name, in) || stem == path.Base(in.Rel) {
			return ""
		}
		return "include/<target>/ holds the public headers named after their slices; move this into the slice whose outcome it serves"
	case cppBarrel(name, in) || cppSnakeRole.MatchString(stem) || pascalRole.MatchString(stem):
		return ""
	}
	return "a C++ file in a slice is <name>_<role> or <Name><Role> (roles: " + strings.Join(Roles, ", ") + "), or the slice's header named after it"
}

// cppBarrel reports whether a header is a slice's API: named after the
// folder it is in, or after a folder beside it.
func cppBarrel(name string, in Dir) bool {
	stem := strings.TrimSuffix(name, path.Ext(name))
	return stem == path.Base(in.Rel) || in.Folders[stem]
}

// cppMixed reports a folder whose files mix snake_case and PascalCase, so
// that a project keeps to one style; the PascalCase file is the one named.
func cppMixed(stem string, in Dir) string {
	if !cppPascal.MatchString(stem) {
		return ""
	}
	for other := range in.Files {
		if !cppExtensions[path.Ext(other)] {
			continue
		}
		otherStem := strings.TrimSuffix(other, path.Ext(other))
		if strings.Contains(otherStem, "_") && snakeCase.MatchString(otherStem) {
			return "this folder mixes snake_case (" + other + ") and PascalCase file names; keep to one style"
		}
	}
	return ""
}
