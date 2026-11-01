package scaffold

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nearmatch"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// likeProject finds the project that smith add --like names.
//
// Parameters: s is the open session; opts hold Like, the project to copy.
// Returns: the project, or nil when --like is not given. It returns an error
// when no project has that name, or the project reports no shape to copy.
func likeProject(s *session.Session, opts AddOptions) (*projectgraph.Project, error) {
	if opts.Like == "" {
		return nil, nil
	}
	if opts.Fresh {
		return nil, fmt.Errorf("--like copies a project and --fresh copies none: use one or the other")
	}
	p := s.Graph.Get(opts.Like)
	if p == nil {
		return nil, fmt.Errorf("--like names %q, which is not a project (projects: %s).%s", opts.Like, strings.Join(s.Graph.Names(), ", "), nearmatch.Suggestion(opts.Like, s.Graph.Names()))
	}
	if len(p.Shape) == 0 {
		return nil, fmt.Errorf("--like names %s, but its plugin reports no build, test or layout to copy. Name a project that has them, or leave --like out", p.Name)
	}
	return p, nil
}
