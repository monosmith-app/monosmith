package release

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
)

// latestChannelTag returns the highest version of the channel that the branch
// checked out contains, and its tag.
func latestChannelTag(git *gitclient.Git, p *projectgraph.Project, c Channel) (Version, string, bool) {
	prefix := TagPrefix(p)
	tags, _ := git.TagsMerged(prefix)
	var best Version
	var bestTag string
	found := false
	for _, t := range tags {
		v, err := ParseVersion(strings.TrimPrefix(t, prefix))
		if _, ok := c.counter(v); err != nil || !ok {
			continue
		}
		if !found || best.Less(v) {
			best, bestTag, found = v, t, true
		}
	}
	return best, bestTag, found
}
