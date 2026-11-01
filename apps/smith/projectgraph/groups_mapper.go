package projectgraph

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// InGroup reports whether the project matches every selector of the group.
func (p *Project) InGroup(g workspace.Group) bool {
	if len(g.Select) == 0 {
		return false
	}
	for _, selector := range g.Select {
		if !p.Matches(strings.TrimSpace(selector)) {
			return false
		}
	}
	return true
}

// GroupSource is how a group is named where a value's origin is shown.
func GroupSource(g workspace.Group) string {
	quoted := make([]string, len(g.Select))
	for i, s := range g.Select {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "smith.toml [[groups]] select = [" + strings.Join(quoted, ", ") + "]"
}

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

// applyGroups layers the [[groups]] that name the project, in file
// order, so a later group wins field by field. A group only changes tasks the
// project already has: it describes how a set of projects runs what they
// share, and a task one of them lacks is not added.
func (p *Project) applyGroups(cfg workspace.Config) {
	for _, g := range cfg.Groups {
		if !p.InGroup(g) {
			continue
		}
		names := make([]string, 0, len(g.Tasks))
		for name := range g.Tasks {
			if _, ok := p.Tasks[name]; ok {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		for _, name := range names {
			p.override(name, g.Tasks[name], GroupSource(g))
		}
	}
}
