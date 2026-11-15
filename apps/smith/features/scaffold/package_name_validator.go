package scaffold

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// packageNameRule is one language's rule for the name a project is published
// or imported under: the pattern, its longest length (0: no limit) and the
// rule in words, which the error names.
type packageNameRule struct {
	what    string
	pattern *regexp.Regexp
	max     int
	words   string
}

// packageNameRules are the rules of the languages whose plugin writes a
// package name into the manifest, by language plugin.
var packageNameRules = map[string]packageNameRule{
	"node": {"npm package name", regexp.MustCompile(`^(@[a-z0-9~-][a-z0-9._~-]*/)?[a-z0-9~-][a-z0-9._~-]*$`), 214,
		"lowercase letters, digits, - . _ ~, an optional @scope/ first, not starting with . or _, at most 214 characters"},
	"rust": {"Cargo crate name", regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`), 64,
		"ASCII letters, digits, - and _, starting with a letter, at most 64 characters"},
	"python": {"Python distribution name", regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9._-]*[A-Za-z0-9])?$`), 0,
		"letters, digits, - . _, starting and ending with a letter or digit (PEP 508)"},
	"go": {"last part of the Go module path", regexp.MustCompile(`^[a-z0-9_~-][a-z0-9._~-]*$`), 0,
		"lowercase letters, digits, - . _ ~, not starting with a dot, no slash"},
	"dotnet": {".NET project id", regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`), 100,
		"C# identifiers joined by dots (Acme.Core), at most 100 characters"},
	"dart": {"pub package name", regexp.MustCompile(`^[a-z_][a-z0-9_]*$`), 0,
		"lowercase letters, digits and _, not starting with a digit"},
	"php": {"Composer package name", regexp.MustCompile(`^([a-z0-9]([_.-]?[a-z0-9]+)*/)?[a-z0-9](([_.]|-{1,2})?[a-z0-9]+)*$`), 0,
		"lowercase vendor/name of letters and digits, single separators - . _ between them; the vendor is the workspace's when left out"},
	"ruby": {"gem name", regexp.MustCompile(`^[a-z][a-z0-9_]*$`), 0,
		"lowercase letters, digits and _, starting with a letter (the Ruby module is made from it)"},
	"jvm": {"Maven artifactId", regexp.MustCompile(`^[A-Za-z0-9_-][A-Za-z0-9._-]*$`), 0,
		"letters, digits, - . _, not starting with a dot"},
	"cpp": {"CMake target name", regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`), 0,
		"letters, digits and _, not starting with a digit (it is the namespace too)"},
	"swift": {"Swift package name", regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`), 0,
		"letters, digits, - and _, starting with a letter"},
	"lua": {"LuaRocks rock name", regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`), 0,
		"lowercase letters, digits, - . _, starting with a letter or digit"},
}

// CheckPackageName refuses a --package-name the language cannot use: a name
// that breaks the language's naming rule, or a language whose plugin writes no
// package name. An empty name is always accepted (the plugin builds one).
//
// Parameters: language is the language plugin's name (rust, node); name is the
// package name the person gave.
// Returns: an error that names the rule, or nil.
func CheckPackageName(language, name string) error {
	if name == "" {
		return nil
	}
	rule, ok := packageNameRules[language]
	if !ok {
		var known []string
		for l := range packageNameRules {
			known = append(known, l)
		}
		sort.Strings(known)
		return fmt.Errorf("--package-name does not apply to a %s project: its plugin writes no package name. It applies to %s", language, strings.Join(known, ", "))
	}
	if !rule.pattern.MatchString(name) || (rule.max > 0 && len(name) > rule.max) {
		return fmt.Errorf("%q is not a valid %s: use %s", name, rule.what, rule.words)
	}
	return nil
}
