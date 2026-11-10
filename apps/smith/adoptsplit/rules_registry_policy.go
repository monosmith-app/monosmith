package adoptsplit

import (
	"path"
	"strings"
)

// rules is every language's declaration. A language joins the split by
// adding its Rules here, and may do so one field at a time: names first, then
// Sort and Split.
var rules = map[string]Rules{
	Node:   nodeRules,
	Go:     goRules,
	Python: pythonRules,
	Rust:   rustRules,
}

// For returns the rules of a language: the zero Rules, which moves everything,
// for one that declares none.
//
// Parameters: ecosystem is one of the ecosystem constants.
func For(ecosystem string) Rules {
	return rules[ecosystem]
}

// StaysAtRoot is true when the top-level name is tool configuration or a
// lockfile that belongs to the workspace for this language.
//
// Parameters: name is a top-level file or folder name, in any case.
func (r Rules) StaysAtRoot(name string) bool {
	lower := strings.ToLower(name)
	return matchesAny(r.Workspace, lower) || matchesAny(r.Lockfiles, lower)
}

func matchesAny(patterns []string, name string) bool {
	for _, p := range patterns {
		if ok, _ := path.Match(p, name); ok {
			return true
		}
	}
	return false
}
