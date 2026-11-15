package languages

import (
	"path"
	"regexp"
	"strings"
)

var (
	dotnetTest = regexp.MustCompile(`^` + pascalName + `Tests?$`)
	pascalCase = regexp.MustCompile(`^` + pascalName + `$`)
	// dotnetGenerated are designer and source-generator outputs.
	dotnetGenerated = regexp.MustCompile(`\.(?:Designer|g|g\.i|AssemblyAttributes|AssemblyInfo)\.(?:cs|fs)$`)
	// dotnetEntry are the files a project's root may hold.
	dotnetEntry = map[string]bool{"Program": true, "GlobalUsings": true, "Usings": true, "AssemblyInfo": true}
)

// dotnetRules follow the .NET naming guidelines: a file is named after the
// type it holds, in PascalCase, and folders are PascalCase namespaces. The
// project's root holds Program.cs; a slice is a folder of <Name><Role>.cs
// files (PlaceOrderUseCase.cs) whose public API is the file named after it
// (Orders/Orders.cs). Tests, in tests/ or a *.Tests project, are
// <Name>Tests.cs, with test doubles in a Testing/ folder; bin/ and obj/
// are build output.
var dotnetRules = Rules{
	Name:   "dotnet",
	Roots:  []string{""},
	Inputs: []string{"**/*.cs", "**/*.fs"},
	Skip: func(name string, _ Dir) bool {
		return name == "bin" || name == "obj" || name == "Properties" || name == "TestResults"
	},
	Folder: func(name string, _ Dir) string {
		if !pascalCase.MatchString(name) && !dotnetTestArea(name) && name != "tests" && name != "test" {
			return "a .NET folder is a PascalCase namespace"
		}
		return ""
	},
	Source: func(name string, _ Dir) bool {
		ext := path.Ext(name)
		return (ext == ".cs" || ext == ".fs") && !dotnetGenerated.MatchString(name)
	},
	File: DotnetFile,
	Production: func(name string, in Dir) bool {
		return in.Rel != "" && !dotnetInTests(in) && !pascalBarrel(name, in)
	},
}

// DotnetFile says why a C# or F# file breaks the rules, or "" when it does
// not.
func DotnetFile(name string, in Dir) string {
	stem := strings.TrimSuffix(name, path.Ext(name))
	switch {
	case dotnetInTests(in):
		if dotnetTest.MatchString(stem) || (pascalTesting(in) && pascalCase.MatchString(stem)) {
			return ""
		}
		return "a test is <Name>Tests.cs; test doubles go in a Testing/ folder"
	case in.Rel == "":
		if dotnetEntry[stem] {
			return ""
		}
		return "the project's root holds Program.cs; move this into the slice whose outcome it serves"
	case pascalBarrel(name, in) || pascalRole.MatchString(stem):
		return ""
	}
	return "a file in a slice is <Name><Role>.cs (roles: " + strings.Join(pascalRoles(), ", ") + "), or the slice's " + path.Base(in.Rel) + path.Ext(name)
}

// dotnetTestArea reports whether a folder holds tests: tests/, test/ or a
// *.Tests project.
func dotnetTestArea(name string) bool {
	return name == "tests" || name == "test" || strings.HasSuffix(name, ".Tests") || strings.HasSuffix(name, ".UnitTests")
}

func dotnetInTests(in Dir) bool {
	for _, part := range strings.Split(in.Rel, "/") {
		if dotnetTestArea(part) {
			return true
		}
	}
	return false
}

// pascalBarrel reports whether a file is a slice's public API in a
// PascalCase language: named after the folder it is in.
func pascalBarrel(name string, in Dir) bool {
	return in.Rel != "" && strings.TrimSuffix(name, path.Ext(name)) == path.Base(in.Rel)
}

// pascalTesting reports whether a test folder is, or is in, a Testing/
// folder of test doubles.
func pascalTesting(in Dir) bool {
	for _, part := range strings.Split(in.Rel, "/") {
		if part == "Testing" || part == "testing" {
			return true
		}
	}
	return false
}
