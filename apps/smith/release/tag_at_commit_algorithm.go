package release

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// TaggedAt is the project's release tag pointing at commit, and its version;
// both are empty when the commit was not released.
func TaggedAt(s *session.Session, p *projectgraph.Project, commit string) (string, string) {
	prefix := TagPrefix(p)
	tags, _ := s.Git.Tags(prefix)
	for _, t := range tags {
		if c, err := s.Git.RevParse("refs/tags/" + t + "^{commit}"); err == nil && c == commit {
			return t, strings.TrimPrefix(t, prefix)
		}
	}
	return "", ""
}
