package releaseassets

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/release"
)

// projectOfTag finds the project with a release build command whose release
// tag is tag, and the version the tag carries. The project with the longest tag
// prefix wins, so a project with a bare tag does not take the tags of others.
// It returns nil when no such project exists.
func projectOfTag(projects []*projectgraph.Project, tag string) (*projectgraph.Project, string) {
	var best *projectgraph.Project
	var bestPrefix string
	for _, p := range projects {
		if !release.BuildsAssetsFromCommand(p) {
			continue
		}
		prefix := release.TagPrefix(p)
		if !strings.HasPrefix(tag, prefix) || len(tag) == len(prefix) {
			continue
		}
		if best == nil || len(prefix) > len(bestPrefix) {
			best, bestPrefix = p, prefix
		}
	}
	if best == nil {
		return nil, ""
	}
	return best, strings.TrimPrefix(tag, bestPrefix)
}
