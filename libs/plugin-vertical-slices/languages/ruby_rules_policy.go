package languages

import (
	"path"
	"regexp"
	"strings"
)

var (
	rubyRoleFile  = regexp.MustCompile(`^` + snakeName + `_(?:` + roleAlternation + `)\.rb$`)
	rubySpecFile  = regexp.MustCompile(`^` + snakeName + `_(?:spec|test)\.rb$`)
	rubySnakeFile = regexp.MustCompile(`^` + snakeName + `\.rb$`)
	// rubyTestSetup are the files RSpec and Minitest load before the tests.
	rubyTestSetup = map[string]bool{"spec_helper.rb": true, "rails_helper.rb": true, "test_helper.rb": true}
)

// rubyRules follow the Ruby style guide: files and folders are snake_case,
// named after the constant they define. lib/<gem>.rb loads the gem; its
// namespace folder lib/<gem>/ holds version.rb and the slices, each a folder
// of <name>_<role>.rb files whose public API is the barrel named after it
// (lib/<gem>/orders/orders.rb, or lib/<gem>/orders.rb beside the folder).
// Specs and tests mirror lib/ as <basename>_spec.rb or <basename>_test.rb,
// with shared setup in support/ or testing/.
var rubyRules = Rules{
	Name:   "ruby",
	Roots:  []string{"lib"},
	Inputs: []string{"lib/**/*.rb", "spec/**/*.rb", "test/**/*.rb"},
	Skip: func(name string, in Dir) bool {
		return in.Rel == "" && name != "lib" && name != "spec" && name != "test"
	},
	Folder: func(name string, _ Dir) string {
		if !snakeCase.MatchString(name) {
			return "a Ruby folder is snake_case, named after the module it holds"
		}
		return ""
	},
	Source: func(name string, in Dir) bool { return in.Rel != "" && strings.HasSuffix(name, ".rb") },
	File:   RubyFile,
	Production: func(name string, in Dir) bool {
		return strings.Count(in.Rel, "/") >= 2 && strings.HasPrefix(in.Rel, "lib/") && !rubyBarrel(name, in)
	},
}

// RubyFile says why a Ruby file breaks the rules, or "" when it does not.
func RubyFile(name string, in Dir) string {
	depth := strings.Count(in.Rel, "/")
	switch area, _, _ := strings.Cut(in.Rel, "/"); {
	case area == "spec" || area == "test":
		if rubySpecFile.MatchString(name) || rubyTestSetup[name] || (rubySupport(in) && rubySnakeFile.MatchString(name)) {
			return ""
		}
		return "a spec is <basename>_spec.rb (a test <basename>_test.rb), mirroring lib/; shared setup goes in support/"
	case in.Rel == "lib":
		if rubySnakeFile.MatchString(name) {
			return ""
		}
		return "a Ruby file is snake_case.rb"
	case depth == 1:
		if name == "version.rb" || rubyBarrel(name, in) {
			return ""
		}
		return "the gem's namespace folder holds version.rb and slice folders; move this into the slice whose outcome it serves"
	case rubyBarrel(name, in) || rubyRoleFile.MatchString(name):
		return ""
	}
	return "a Ruby file in a slice is <name>_<role>.rb (roles: " + strings.Join(Roles, ", ") + "), or the slice's barrel " + path.Base(in.Rel) + ".rb"
}

// rubyBarrel reports whether a file is a slice's public API: named after its
// folder, or after a folder beside it.
func rubyBarrel(name string, in Dir) bool {
	module := strings.TrimSuffix(name, ".rb")
	return module == path.Base(in.Rel) || in.Folders[module]
}

// rubySupport reports whether a spec folder holds shared setup: support/ or
// testing/.
func rubySupport(in Dir) bool {
	for _, part := range strings.Split(in.Rel, "/") {
		if part == "support" || part == "testing" {
			return true
		}
	}
	return false
}
