package imports

import (
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"
)

var (
	pyAnyFrom      = regexp.MustCompile(`(?m)^[ \t]*from[ \t]+([\w.]+)[ \t]+import\b`)
	rustAnyUse     = regexp.MustCompile(`(?m)^[ \t]*(?:pub(?:\([\w ]+\))?[ \t]+)?(?:use|extern[ \t]+crate)[ \t]+:?:?(\w+(?:::\w+)*)`)
	rubyAnyRequire = regexp.MustCompile(`(?m)^[ \t]*require[ \t(]+['"]([^'"\n]+)['"]`)
	cppAnyInclude  = regexp.MustCompile(`(?m)^[ \t]*#[ \t]*include[ \t]+[<"]([^>"\n]+)[>"]`)
)

// All lists every import of one source file as the language spells it, the
// project's own and every library's: a Go import path, a Python module, a
// JVM or .NET name, a PHP namespace, a Dart URI, a Rust path, a Ruby
// require, a C++ include. The vendor rule matches it against the names a
// project lists. A language the checker cannot read gives nil.
//
// Parameters: language is the project's language (go, python, jvm, ...);
// text is the file's source.
// Returns the imports in the order they appear.
func All(language, text string) []string {
	var out []string
	add := func(re *regexp.Regexp, group int) {
		for _, m := range re.FindAllStringSubmatch(text, -1) {
			out = append(out, m[group])
		}
	}
	switch canonical(language) {
	case "go":
		file, _ := parser.ParseFile(token.NewFileSet(), "", text, parser.ImportsOnly)
		if file == nil {
			return nil
		}
		for _, spec := range file.Imports {
			if target, err := strconv.Unquote(spec.Path.Value); err == nil {
				out = append(out, target)
			}
		}
	case "python":
		add(pyAnyFrom, 1)
		for _, m := range pyImport.FindAllStringSubmatch(text, -1) {
			for _, item := range strings.Split(m[1], ",") {
				if fields := strings.Fields(item); len(fields) > 0 {
					out = append(out, fields[0])
				}
			}
		}
	case "jvm":
		add(jvmImport, 1)
	case "dotnet":
		add(dotnetUsing, 1)
	case "php":
		add(phpUse, 1)
	case "dart":
		add(dartSpec, 1)
	case "rust":
		add(rustAnyUse, 1)
	case "ruby":
		add(rubyAnyRequire, 1)
	case "cpp":
		add(cppAnyInclude, 1)
	}
	return out
}
