package release

import (
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// ResumeReason is the reason of an item that only finishes a release.
const ResumeReason = "tagged, but its GitHub Release is missing"

// UnfinishedReleases finds the projects whose latest release tag exists but
// whose GitHub Release does not: a release that stopped after the tag, whose
// re-run on a fresh checkout (CI) has no journal and no new commits to plan
// from. Each is an item for a tagged release, which publishes nothing to any
// registry and only creates the Release and attaches its assets. A project
// without a release tag is left out, and so is every project when GitHub
// Releases are off.
func UnfinishedReleases(s *session.Session, hasRelease func(tag string) (bool, error)) ([]Item, error) {
	if !s.Config.Release.GitHubReleases {
		return nil, nil
	}
	order, err := s.Graph.Order()
	if err != nil {
		return nil, err
	}
	var items []Item
	for _, name := range order {
		p := s.Graph.Get(name)
		if !p.Releases() {
			continue
		}
		latest, tag, released := LatestTag(s.Git, p)
		if !released {
			continue
		}
		found, err := hasRelease(tag)
		if err != nil {
			return nil, fmt.Errorf("GitHub Release %s: %w", tag, err)
		}
		if !found {
			items = append(items, Item{Project: name, Plugin: p.Plugin, Current: latest.String(), Next: latest.String(), Level: "resume", Tag: tag, Reason: ResumeReason})
		}
	}
	return items, nil
}

// UnfinishedFromGitHub is UnfinishedReleases asking GitHub. Without a token or
// a repository it finds nothing (a local run has no Release to look for); a
// lookup that fails is said, not fatal: the plan stays empty.
func UnfinishedFromGitHub(s *session.Session, warn io.Writer) []Item {
	client, err := githubrelease.FromEnvironment()
	if err != nil {
		return nil
	}
	repo, err := Repository(s)
	if err != nil {
		return nil
	}
	items, err := UnfinishedReleases(s, func(tag string) (bool, error) {
		_, found, err := client.ReleaseByTag(repo, tag)
		return found, err
	})
	if err != nil {
		fmt.Fprintf(warn, "warning: could not look for unfinished releases: %v\n", err)
		return nil
	}
	return items
}
