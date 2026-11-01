package release

import (
	"slices"
	"strings"
	"sync"

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
	var releasing []*projectgraph.Project
	for _, p := range s.Graph.Projects {
		if p.Releases() {
			releasing = append(releasing, p)
		}
	}
	// The registry lookups wait on the network, so they run together; the
	// plugin's declared concurrency bounds how many it takes at once. Each
	// goroutine fills its own slot, so the order is the graph's.
	subjects := make([]releaselineage.Subject, len(releasing))
	var wg sync.WaitGroup
	var mu sync.Mutex
	hung := map[string]bool{} // plugins whose registry did not answer in time
	turn := make(chan struct{}, lookupsAtOnce)
	for i, p := range releasing {
		subjects[i] = releaselineage.Subject{Project: p.Name, Prefix: TagPrefix(p), Alternatives: alternatives(p), Explicit: p.TagPrefixSet, Tags: tags}
		wg.Add(1)
		go func() {
			defer wg.Done()
			turn <- struct{}{}
			defer func() { <-turn }()
			mu.Lock()
			skip := hung[p.Plugin]
			mu.Unlock()
			if skip {
				// An earlier lookup of this plugin found its registry silent:
				// do not make every other project wait the same time.
				subjects[i].Unreachable, subjects[i].Detail = true, protocol.RegistryUnreachableDetail
				return
			}
			var res protocol.PublishedResult
			if ok, err := s.Plugins.Call(p.Plugin, protocol.MethodPublished, protocol.PublishedParams{Project: p.Project}, &res, lookupCallTimeout); ok {
				if err != nil {
					res = protocol.PublishedResult{Unreachable: true, Detail: err.Error()}
				}
				subjects[i].Published, subjects[i].Unreachable, subjects[i].Registry, subjects[i].Detail = res.Version, res.Unreachable, res.Registry, res.Detail
				if res.Unreachable && res.Detail == protocol.RegistryUnreachableDetail {
					mu.Lock()
					hung[p.Plugin] = true
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	return append(findings, releaselineage.Assess(subjects)...)
}

// lookupsAtOnce is how many registry lookups the check has in flight at once.
const lookupsAtOnce = 8

// lookupCallTimeout is the host's backstop for one registry lookup. A plugin
// ends a silent lookup after protocol.RegistryLookupTimeout, so the backstop,
// which kills the plugin, is reached only by a plugin that does not.
const lookupCallTimeout = 4 * protocol.RegistryLookupTimeout

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
