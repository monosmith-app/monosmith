package doctor

import (
	"strconv"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// kindHintChecks are suggestions about a project's kind: a lib nothing uses,
// a package nothing publishes. Each points at `smith kind`. They are
// warnings, so they never count as failures, and nothing is reported when
// there is nothing to say.
func kindHintChecks(s *session.Session) []protocol.Finding {
	var out []protocol.Finding
	now := time.Now()
	for _, p := range s.Graph.Projects {
		switch p.Kind {
		case protocol.ProjectLib:
			dependents := len(s.Graph.Dependents(p.Name))
			if dependents > 0 || len(s.Graph.Projects) < 2 {
				continue
			}
			if f, ok := UnusedLibHint(p.Name, dependents, len(s.Graph.Projects)-1, addedAt(s, p.Root), now); ok {
				out = append(out, f)
			}
		case protocol.ProjectPackage:
			// A workspace that releases everything by tag (release registry
			// none) chose that for all of its packages; only the odd one out,
			// named in its own settings or manifest, is worth a word.
			own := s.Config.Projects[p.Name].Registry == workspace.NoRegistry
			never := (own || p.PublishTo == "none") && s.Config.Release.Registry != workspace.NoRegistry
			if f, ok := UnpublishedPackageHint(p.Name, never); ok {
				out = append(out, f)
			}
		}
	}
	return out
}

// addedAt is when the project's directory first had a file committed, or the
// zero time when git does not know (a project that was never committed).
func addedAt(s *session.Session, root string) time.Time {
	raw, err := s.Git.Run("log", "--diff-filter=A", "--format=%ct", "--", root)
	if err != nil {
		return time.Time{}
	}
	stamps := strings.Fields(raw)
	if len(stamps) == 0 {
		return time.Time{}
	}
	seconds, err := strconv.ParseInt(stamps[len(stamps)-1], 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(seconds, 0)
}
