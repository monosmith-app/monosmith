package languages

import (
	"path"
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

// WholeNameOnly are banned words that are honest inside a compound
// (user_data_contract.go, account_manager_policy.py): only a file whose name,
// before its role, is the word is reported.
var WholeNameOnly = map[string]bool{"data": true, "manager": true}

// Zones are the folders an app's source root holds; every slice is in one.
// shared is a junk drawer anywhere else.
var Zones = []string{"features", "integrations", "shared"}

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
	if stem := strings.ToLower(nameBeforeRole(name)); WholeNameOnly[stem] {
		return stem
	}
	for _, part := range strings.FieldsFunc(strings.ToLower(name), func(r rune) bool { return r == '_' || r == '.' || r == '-' }) {
		if ForbiddenFolders[part] && !WholeNameOnly[part] {
			return part
		}
	}
	return ""
}

// RoleOf is the role a file's name ends in, in snake_case (use_case), or ""
// when it ends in none: place_order_use_case.go, PlaceOrderUseCase.kt. A
// test's marker (_test, test_, Test, Tests, _spec, Spec) is not part of it.
func RoleOf(name string) string {
	stem := withoutTestMarker(strings.TrimSuffix(name, path.Ext(name)))
	for _, role := range rolesLongestFirst {
		if strings.HasSuffix(stem, "_"+role) || strings.HasSuffix(stem, pascal(role)) && stem != pascal(role) {
			return role
		}
	}
	return ""
}

// withoutTestMarker takes a test's marker off a file stem.
func withoutTestMarker(stem string) string {
	for _, suffix := range []string{"_test", "_spec", "Tests", "Test", "Spec"} {
		if cut, ok := strings.CutSuffix(stem, suffix); ok && cut != "" {
			return cut
		}
	}
	if cut, ok := strings.CutPrefix(stem, "test_"); ok && cut != "" {
		return cut
	}
	return stem
}

// nameBeforeRole is a file's name without its extension, test marker and role.
func nameBeforeRole(name string) string {
	stem := withoutTestMarker(strings.TrimSuffix(name, path.Ext(name)))
	if role := RoleOf(name); role != "" {
		stem = strings.TrimSuffix(strings.TrimSuffix(stem, "_"+role), pascal(role))
	}
	return stem
}

// rolesLongestFirst lets use_case win over case-like shorter endings.
var rolesLongestFirst = func() []string {
	out := append([]string(nil), Roles...)
	sort.Slice(out, func(i, j int) bool { return len(out[i]) > len(out[j]) })
	return out
}()

// pascal spells one role in PascalCase: use_case -> UseCase.
func pascal(role string) string {
	var b strings.Builder
	for _, word := range strings.Split(role, "_") {
		b.WriteString(strings.ToUpper(word[:1]) + word[1:])
	}
	return b.String()
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
	// Roots are where the project's sources start, tried in order, relative
	// to the project root ("" is the root itself). From the first that
	// exists, a chain of single folders is followed down (a JVM package
	// path, lib/<gem>): the zones of an app sit there.
	Roots []string
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
