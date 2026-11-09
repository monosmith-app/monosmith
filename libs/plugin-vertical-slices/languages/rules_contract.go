package languages

import (
	"regexp"
	"sort"
	"strings"
)

// Roles are the role suffixes a file may carry, in snake_case; a PascalCase
// language spells them UseCase, Handler and so on.
var Roles = []string{"handler", "use_case", "algorithm", "policy", "model", "contract", "mapper", "validator", "repository", "client", "store", "middleware", "error", "config", "enum"}

// ForbiddenFolders are junk drawers, in lowercase (Helpers/ is one too): a
// folder is named for the outcome it serves.
var ForbiddenFolders = map[string]bool{"util": true, "utils": true, "helper": true, "helpers": true, "common": true, "shared": true, "manager": true, "processor": true, "data": true, "misc": true}

// ToolMandatedNames are file names a test tool loads by name, so a banned word
// in them is not the project's choice: RSpec's and Minitest's setup files. The
// file-name check skips them.
var ToolMandatedNames = rubyTestSetup

// BannedWordIn returns the first part of a file name, split on _ . and -, that
// is a ForbiddenFolders word (compared in lowercase), or "" when there is
// none. The name of a tool-mandated file has none.
func BannedWordIn(name string) string {
	if ToolMandatedNames[name] {
		return ""
	}
	for _, part := range strings.FieldsFunc(strings.ToLower(name), func(r rune) bool { return r == '_' || r == '.' || r == '-' }) {
		if ForbiddenFolders[part] {
			return part
		}
	}
	return ""
}

// Dir is a folder being checked.
type Dir struct {
	// Rel is its path from the project root, with forward slashes; "" is
	// the root.
	Rel string
	// Files and Folders are the names of the files and folders in it.
	Files, Folders map[string]bool
	// Parent is the folder it is in; nil at the root.
	Parent *Dir
}

// Rules are one language's vertical-slice layout. Every function is
// optional except File.
type Rules struct {
	// Name is the language as its plugin names it (go, python, dart).
	Name string
	// Inputs are the globs the check reads, which key its cache.
	Inputs []string
	// Skip reports whether a folder holds nothing the rules check.
	Skip func(name string, in Dir) bool
	// Folder says why a folder's name breaks the rules, or "".
	Folder func(name string, in Dir) string
	// Source reports whether the rules look at a file at all.
	Source func(name string, in Dir) bool
	// File says why a source file breaks the rules, or "".
	File func(name string, in Dir) string
	// Production reports whether a source file counts toward its slice's
	// size: code, not a test or a barrel.
	Production func(name string, in Dir) bool
}

var all = []Rules{goRules, pythonRules, dartRules, rustRules, rubyRules, cppRules, dotnetRules, jvmRules, phpRules}

// For returns a language's rules.
func For(language string) (Rules, bool) {
	for _, r := range all {
		if r.Name == language {
			return r, true
		}
	}
	return Rules{}, false
}

// Names lists the languages that have rules, sorted.
func Names() []string {
	var names []string
	for _, r := range all {
		names = append(names, r.Name)
	}
	sort.Strings(names)
	return names
}

var (
	roleAlternation = strings.Join(Roles, "|")
	snakeName       = `[a-z][a-z0-9]*(?:_[a-z0-9]+)*`
	snakeCase       = regexp.MustCompile(`^` + snakeName + `$`)
	// pascalName is a PascalCase name; pascalRole one ending in a role
	// (PlaceOrderUseCase), for languages that name files after classes.
	pascalName = `[A-Z][A-Za-z0-9]*`
	pascalRole = regexp.MustCompile(`^` + pascalName + `(?:` + strings.Join(pascalRoles(), "|") + `)$`)
)

// pascalRoles spells the roles in PascalCase: use_case → UseCase.
func pascalRoles() []string {
	out := make([]string, len(Roles))
	for i, role := range Roles {
		var b strings.Builder
		for _, word := range strings.Split(role, "_") {
			b.WriteString(strings.ToUpper(word[:1]) + word[1:])
		}
		out[i] = b.String()
	}
	return out
}

// inTesting reports whether a test folder is, or is in, a testing/ folder of
// test doubles.
func inTesting(in Dir) bool {
	for _, part := range strings.Split(in.Rel, "/") {
		if part == "testing" {
			return true
		}
	}
	return false
}
