package release

import (
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Channel is the release channel of the branch being released: the line of
// prerelease versions <Major>.0.0-<ID>.<n> it publishes (see
// workspace.ReleaseChannel).
type Channel struct {
	ID    string
	Major int
}

// ChannelOf returns the channel that releases from branch, or nil when the
// branch releases stable versions.
func ChannelOf(cfg workspace.ReleaseSection, branch string) *Channel {
	c, ok := cfg.ChannelFor(branch)
	if !ok {
		return nil
	}
	return &Channel{ID: c.Prerelease, Major: c.Major}
}

// counter returns n of a version <Major>.x.y-<ID>.<n> of this channel.
func (c Channel) counter(v Version) (int, bool) {
	rest, ok := strings.CutPrefix(v.Pre, c.ID+".")
	if !ok || v.Major != c.Major {
		return 0, false
	}
	n, err := strconv.Atoi(rest)
	return n, err == nil
}

// Next is the version after last, the channel's latest release when onChannel
// is true. The first release of a channel is <Major>.0.0-<ID>.1, whatever
// stable versions exist; later ones raise the counter and keep the numbers, so
// a breaking change inside a channel does not start a new major.
func (c Channel) Next(last Version, onChannel bool) Version {
	if !onChannel {
		return Version{Major: c.Major, Pre: c.ID + ".1"}
	}
	n, _ := c.counter(last)
	last.Pre = c.ID + "." + strconv.Itoa(n+1)
	return last
}

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
