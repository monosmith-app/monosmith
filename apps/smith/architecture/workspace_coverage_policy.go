package architecture

import (
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// Member is one project of the workspace, as far as an architecture cares: its
// name and the language it is written in.
type Member struct {
	Name, Language string
}

// Coverage says which projects of a workspace an architecture would check and
// which it would skip because their language is not one it applies to.
type Coverage struct {
	Checked, Skipped []Member
}

// Cover says what choosing the architecture does to the members. The choice of
// no architecture checks nothing and skips nothing: it is not a gap.
func (a Architecture) Cover(members []Member) Coverage {
	var c Coverage
	if a.Table == "" {
		return c
	}
	declared := a.AppliesTo()
	for _, m := range members {
		if projectgraph.CoversLanguage(declared, m.Language) {
			c.Checked = append(c.Checked, m)
		} else {
			c.Skipped = append(c.Skipped, m)
		}
	}
	return c
}

// CoversNone is true when the workspace has projects and the architecture
// would check none of them.
func (c Coverage) CoversNone() bool { return len(c.Checked) == 0 && len(c.Skipped) > 0 }
