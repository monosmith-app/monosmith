package release

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/features/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
)

// PromoteFindings compares a stable plan with what the release channels
// without a major last released: a planned version that no channel released
// as a prerelease (6.2.0 after 6.2.0-beta.N) was not tested there. It looks
// only on the release branch, with [release] promote "warn" or "strict".
//
// Parameters: s is the open session; plan is the plan about to be released.
// Returns one sentence per project whose planned version differs, or nil.
func PromoteFindings(s *session.Session, plan Plan) []string {
	if plan.Channel != "" || s.Config.Release.Promote == "off" || ChannelOf(s.Config.Release, PlanBranch(s)) != nil {
		return nil
	}
	return promoteProblems(s.Git, s.Graph, plan, s.Config.Release.Channels)
}

// PromoteRefusals is PromoteFindings when [release] promote is "strict", the
// findings a release refuses; nil otherwise.
//
// Parameters: s is the open session; plan is the plan about to be released.
// Returns the findings that stop the release, or nil.
func PromoteRefusals(s *session.Session, plan Plan) []string {
	if s.Config.Release.Promote != "strict" {
		return nil
	}
	return PromoteFindings(s, plan)
}

// promoteWarnings is PromoteFindings when [release] promote is "warn".
func promoteWarnings(s *session.Session, plan Plan) []string {
	if s.Config.Release.Promote != "warn" {
		return nil
	}
	return PromoteFindings(s, plan)
}

// promoteProblems checks each planned project against the highest prerelease
// that a channel without a major released and that HEAD contains.
func promoteProblems(git *gitclient.Git, graph *projectgraph.Graph, plan Plan, channels []workspace.ReleaseChannel) []string {
	var fromNext []workspace.ReleaseChannel
	for _, c := range channels {
		if c.Major == 0 {
			fromNext = append(fromNext, c)
		}
	}
	if len(fromNext) == 0 {
		return nil
	}
	var problems []string
	for _, item := range plan.Items {
		p := graph.Get(item.Project)
		next, err := ParseVersion(item.Next)
		if p == nil || err != nil {
			continue
		}
		var best Version
		var bestChannel workspace.ReleaseChannel
		found := false
		for _, c := range fromNext {
			v, _, ok := latestChannelTag(git, p, Channel{ID: c.Prerelease})
			if ok && (!found || best.Less(v)) {
				best, bestChannel, found = v, c, true
			}
		}
		switch {
		case !found:
			problems = append(problems, fmt.Sprintf("%s: %s was never released on a channel (%s), so no channel build tested it. Release it from %s first, then merge", item.Project, item.Next, channelBranches(fromNext), channelBranches(fromNext)))
		case !SameRelease(next, best):
			problems = append(problems, fmt.Sprintf("%s: this release plans %s, but the %s channel last released %s, so %s was not tested there. Release %s again, then merge it", item.Project, item.Next, bestChannel.Branch, best, item.Next, bestChannel.Branch))
		}
	}
	return problems
}

// channelBranches names the branches of channels, joined with " or ".
func channelBranches(channels []workspace.ReleaseChannel) string {
	out := ""
	for i, c := range channels {
		if i > 0 {
			out += " or "
		}
		out += c.Branch
	}
	return out
}
