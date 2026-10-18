package release

import (
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/releaselineage"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckLineage checks every releasable project's tags against the names it
// was tagged under before and against its registry. With fetch, it first
// brings the release tags up to date from the release remote, so the answer
// is not as old as the clone; a failed fetch is a warning.
func CheckLineage(s *session.Session, fetch bool) []releaselineage.Finding {
	var findings []releaselineage.Finding
	remote := s.Config.Release.Remote
	if fetch && s.Git.HasRemote(remote) {
		if _, err := s.Git.Run("fetch", remote, "--tags", "--force", "--quiet"); err != nil {
			findings = append(findings, releaselineage.Finding{Check: releaselineage.CheckResolve, Message: "could not fetch the release tags from " + remote + ", so what follows is as old as this clone; run git fetch --tags and try again."})
		}
	}
	tags, _ := s.Git.Tags("")
	var subjects []releaselineage.Subject
	for _, p := range s.Graph.Projects {
		if !p.Releases() {
			continue
		}
		subject := releaselineage.Subject{Project: p.Name, Prefix: TagPrefix(p), Alternatives: alternatives(p), Explicit: p.TagPrefixSet, Tags: tags}
		var res protocol.PublishedResult
		if ok, err := s.Plugins.Call(p.Plugin, protocol.MethodPublished, protocol.PublishedParams{Project: p.Project}, &res, time.Minute); ok {
			if err != nil {
				res = protocol.PublishedResult{Unreachable: true, Detail: err.Error()}
			}
			subject.Published, subject.Unreachable, subject.Registry, subject.Detail = res.Version, res.Unreachable, res.Registry, res.Detail
		}
		subjects = append(subjects, subject)
	}
	return append(findings, releaselineage.Assess(subjects)...)
}

// alternatives are the names a project's tags may carry from before: its
// exported package names, scoped and unscoped, and its own name.
func alternatives(p *projectgraph.Project) []string {
	var out []string
	add := func(prefix string) {
		if !slices.Contains(out, prefix) {
			out = append(out, prefix)
		}
	}
	for _, e := range p.Exports {
		if strings.ContainsAny(e, " \t") {
			continue
		}
		add(e + "@")
		if _, unscoped, scoped := strings.Cut(e, "/"); scoped && strings.HasPrefix(e, "@") {
			add(unscoped + "@")
		}
	}
	add(p.Name + "@")
	return out
}
