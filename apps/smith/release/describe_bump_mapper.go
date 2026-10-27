package release

import (
	"fmt"
	"strings"
)

// maxWhyDescription keeps a commit's description readable in a table cell.
const maxWhyDescription = 56

// Why says in a few words which commit decided the release's version bump:
// the breaking change if there is one, else the first feature, else the first
// fix, with how many other releasing commits came with it. A release forced by
// a version(...) directive or caused by a dependency gives its Reason.
func (i Item) Why() string {
	if i.Forced || len(i.Changes) == 0 {
		return i.Reason
	}
	decisive := i.Changes[0]
	rank := func(c Change) int {
		switch {
		case c.Breaking:
			return 3
		case c.Type == "feat":
			return 2
		}
		return 1
	}
	for _, c := range i.Changes[1:] {
		if rank(c) > rank(decisive) {
			decisive = c
		}
	}
	kind := decisive.Type
	if len(decisive.Scopes) > 0 {
		kind += "(" + strings.Join(decisive.Scopes, ",") + ")"
	}
	if decisive.Breaking {
		kind += "!"
	}
	desc := decisive.Description
	if r := []rune(desc); len(r) > maxWhyDescription {
		desc = string(r[:maxWhyDescription-1]) + "…"
	}
	why := fmt.Sprintf("%s: %s", kind, desc)
	if more := len(i.Changes) - 1; more > 0 {
		why += fmt.Sprintf(" (+%d more)", more)
	}
	return why
}
