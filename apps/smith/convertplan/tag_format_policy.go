package convertplan

import (
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// ScopedTagFormat is the [release] tag_format convert writes when the
// workspace's tags are scoped: "@acme/kit@" for the project kit.
const ScopedTagFormat = "{scope}/{name}@"

// tagChoice is the release tag prefix planned for one project that smith
// prefixes with its name.
type tagChoice struct {
	// project is the smith name.
	project string
	// prefix is the prefix that continues the project's tags.
	prefix string
	// fresh means the project has no tag yet under any name, so any prefix
	// continues it.
	fresh bool
}

// chooseTagFormat decides whether the workspace's tags follow a common
// format, so [release] tag_format replaces a tag_prefix on each project. It
// returns the format and the prefixes the other projects must still set
// (the exceptions), or "" and nil when no format pays off: it needs a scope,
// and at least two projects that follow the format, and more than those that do not.
func chooseTagFormat(scope string, choices []tagChoice) (format string, exceptions map[string]string) {
	if scope == "" {
		return "", nil
	}
	follows, other := 0, map[string]string{}
	for _, c := range choices {
		switch {
		case c.fresh:
		case c.prefix == workspace.ExpandTagFormat(ScopedTagFormat, scope, c.project):
			follows++
		default:
			other[c.project] = c.prefix
		}
	}
	if follows < 2 || follows <= len(other) {
		return "", nil
	}
	return ScopedTagFormat, other
}
