package projectgraph

import (
	"slices"
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// groupSettings adds the tags of the [[groups]] that name the project to its
// tags, in file order, so a later group can select by a tag an earlier one
// gave. It returns those tags and the release table of the last matching group
// that sets one, for the project that sets none of its own.
func (p *Project) groupSettings(cfg workspace.Config) ([]string, *workspace.ProjectRelease) {
	var tags []string
	var release *workspace.ProjectRelease
	for _, g := range cfg.Groups {
		if !p.InGroup(g) {
			continue
		}
		for _, tag := range g.Tags {
			if !slices.Contains(p.Tags, tag) {
				p.Tags = append(p.Tags, tag)
				tags = append(tags, tag)
			}
		}
		if g.Release != nil {
			release = g.Release
		}
	}
	sort.Strings(p.Tags)
	return tags, release
}
