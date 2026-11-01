package scaffold

import (
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/dialect"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// followNote says, for the plan shown by --dry-run, which project or siblings
// the new project will follow and on what: "; follows lib core (test.runner =
// vitest, ...)" or "; follows its 3 sibling libs (...)". It is empty when the
// project follows nothing.
func followNote(s *session.Session, opts AddOptions, kind protocol.ProjectKind, plugin string) string {
	if opts.Fresh {
		return ""
	}
	if like, err := likeProject(s, opts); err == nil && like != nil {
		keys := make([]string, 0, len(like.Shape))
		for k, v := range like.Shape {
			keys = append(keys, k+" = "+v)
		}
		sort.Strings(keys)
		return "; follows " + like.Name + " (" + strings.Join(keys, ", ") + ")"
	}
	var shapes []map[string]string
	for _, p := range s.Graph.Projects {
		if p.Plugin == plugin && p.Kind == kind && p.Shape != nil {
			shapes = append(shapes, p.Shape)
		}
	}
	if lines := dialect.Summary(shapes, string(kind)); len(lines) > 0 {
		return "; follows its sibling " + string(kind) + "s (" + strings.Join(lines, ", ") + ")"
	}
	return ""
}
