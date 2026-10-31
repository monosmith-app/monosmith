package scaffold

import (
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/dialect"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// siblingDialect finds the dialect a new project of the kind follows: the
// shapes of the workspace's projects of the same kind and language, counted by
// the dialect package, and the shape keys the plugins of its stack apply.
func siblingDialect(s *session.Session, opts AddOptions, kind protocol.ProjectKind, language *host.Client, st protocol.Stack) dialect.Dialect {
	var shapes []map[string]string
	for _, p := range s.Graph.Projects {
		if p.Plugin == language.Info.Plugin.Name && p.Kind == kind && p.Shape != nil {
			shapes = append(shapes, p.Shape)
		}
	}
	follows := map[string]bool{}
	plugins := []*host.Client{language}
	for _, name := range st.Plugins() {
		if c := s.Plugins.Get(name); c != nil {
			plugins = append(plugins, c)
		}
	}
	for _, c := range plugins {
		for _, key := range c.Info.Follows {
			follows[key] = true
		}
	}
	var in io.Reader
	interactive := false
	if opts.Recipe != nil {
		in, interactive = opts.Recipe.In, opts.Recipe.Interactive
	}
	return dialect.Resolve(dialect.Request{
		Siblings: shapes, Kind: string(kind), Follows: follows,
		Interactive: interactive, In: in, Out: opts.Out,
	})
}
