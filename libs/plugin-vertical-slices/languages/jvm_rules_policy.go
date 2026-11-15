package languages

import (
	"path"
	"regexp"
	"strings"
)

var (
	jvmTest    = regexp.MustCompile(`^` + pascalName + `(?:Tests?|IT)$`)
	jvmEntry   = regexp.MustCompile(`^(?:Main|App|` + pascalName + `(?:Application|Activity))$`)
	jvmPackage = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	// jvmSourceSets are Gradle's and Maven's source sets that hold tests.
	jvmTestSets = map[string]bool{"test": true, "androidTest": true, "testFixtures": true, "integrationTest": true}
	// jvmResources hold no classes.
	jvmResources = map[string]bool{"res": true, "resources": true, "assets": true, "generated": true}
)

// jvmRules follow the Java and Kotlin conventions: a file is named after
// its class in PascalCase, with the role as a suffix (PlaceOrderUseCase.kt),
// and packages (folders) are lowercase. Sources live in src/<set>/<java or
// kotlin>/; a slice is a package of <Name><Role> files, and entry points
// (Main, Application, an Android activity) sit beside them. Tests in
// src/test/ are <Name>Test, with test doubles in a testing package.
var jvmRules = Rules{
	Name:   "jvm",
	Roots:  []string{"src/main/kotlin", "src/main/java", "src/main/scala", "src"},
	Inputs: []string{"src/**/*.kt", "src/**/*.java"},
	Skip: func(name string, in Dir) bool {
		return (in.Rel == "" && name != "src") || jvmResources[name]
	},
	Folder: func(name string, in Dir) string {
		// src/<set>/ is a source set (main, test, androidTest), named by the
		// build tool.
		if in.Rel == "src" || jvmPackage.MatchString(name) {
			return ""
		}
		return "a package is lowercase"
	},
	Source: func(name string, _ Dir) bool {
		return strings.HasSuffix(name, ".kt") || strings.HasSuffix(name, ".java")
	},
	File: JvmFile,
	Production: func(name string, in Dir) bool {
		return !jvmInTests(in) && !jvmEntry.MatchString(strings.TrimSuffix(name, path.Ext(name)))
	},
}

// JvmFile says why a Kotlin or Java file breaks the rules, or "" when it
// does not.
func JvmFile(name string, in Dir) string {
	stem := strings.TrimSuffix(name, path.Ext(name))
	switch {
	case jvmInTests(in):
		if jvmTest.MatchString(stem) || (inTesting(in) && pascalCase.MatchString(stem)) {
			return ""
		}
		return "a test is <Name>Test; test doubles go in a testing package"
	case jvmEntry.MatchString(stem) || pascalRole.MatchString(stem):
		return ""
	}
	return "a class in a slice is <Name><Role> (roles: " + strings.Join(pascalRoles(), ", ") + "); entry points are Main or <Name>Application"
}

// jvmInTests reports whether a folder is in a test source set.
func jvmInTests(in Dir) bool {
	parts := strings.Split(in.Rel, "/")
	return len(parts) > 1 && parts[0] == "src" && jvmTestSets[parts[1]]
}
