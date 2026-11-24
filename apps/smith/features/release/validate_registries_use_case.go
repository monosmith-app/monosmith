package release

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/features/npmbinaries"
	"github.com/monosmith-app/monosmith/apps/smith/features/releaserehearsal"
	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// RegistryTargets lists the registries a plan publishes an app's binaries
// to, in publishing order: npm, PyPI, NuGet, RubyGems, per package.
func RegistryTargets(s *session.Session, plan Plan) []RegistryTarget {
	var out []RegistryTarget
	for _, item := range plan.Items {
		p := s.Graph.Get(item.Project)
		if p == nil || p.Release == nil || Native(p) || PublishesToNoRegistry(p) || s.Git.TagCommit(item.Tag) != "" {
			continue // a tagged package is finished: only its GitHub Release resumes
		}
		r := p.Release
		add := func(registry string, names ...string) {
			out = append(out, RegistryTarget{Project: item.Project, Registry: registry, Names: names, Version: item.Next})
		}
		if len(r.NPM) > 0 {
			add("npm", r.NPM...)
			if reg := NPMRegistry(p); reg != nil {
				out[len(out)-1].URL, out[len(out)-1].Token = npmTarget(s.Root, reg, os.Getenv)
			}
		}
		if r.PyPI != "" && !r.SkipsPrerelease("pypi", item.Next) {
			add("PyPI", r.PyPI)
		}
		if r.NuGet != "" && !r.SkipsPrerelease("nuget", item.Next) {
			add("NuGet", r.NuGet)
		}
		if r.RubyGems != "" && !r.SkipsPrerelease("rubygems", item.Next) {
			add("RubyGems", r.RubyGems)
		}
	}
	return out
}

// npmTarget is the registry an app's npm packages publish to and the token
// npm signs in to it with, read where the publish reads it (see
// npmbinaries.Npmrc): NPM_CONFIG_REGISTRY overrides the URL, as it does for
// npm; the registry's own variable comes first, then the workspace .npmrc's
// line for it (Azure's npmAuthenticate@0 writes one), then NODE_AUTH_TOKEN.
func npmTarget(root string, reg *protocol.Registry, env func(string) string) (registryURL, token string) {
	registryURL = reg.NPM
	if v := env("NPM_CONFIG_REGISTRY"); v != "" {
		registryURL = v
	}
	_, tokenEnv := npmbinaries.RegistryOf(reg)
	if tokenEnv != "" {
		if token = env(tokenEnv); token != "" {
			return registryURL, token
		}
	}
	if rc, err := os.ReadFile(filepath.Join(root, ".npmrc")); err == nil {
		if token, _ = releaserehearsal.NpmrcToken(string(rc), registryURL, env); token != "" {
			return registryURL, token
		}
	}
	return registryURL, env("NODE_AUTH_TOKEN")
}

// ValidateRegistries asks every registry of a plan whether it would accept
// the release, before anything is published to any: a bad credential on the
// last registry is found here, not after the first ones published. It
// returns every finding; the blocking ones stop the release.
func ValidateRegistries(targets []RegistryTarget, probe *RegistryProbe) []RegistryFinding {
	// The registries are asked at the same time; the findings keep the order of
	// the targets. The probe holds no state that a call changes.
	found := make([][]RegistryFinding, len(targets))
	var wg sync.WaitGroup
	for i, t := range targets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			found[i] = probe.Check(t)
		}()
	}
	wg.Wait()
	var out []RegistryFinding
	for _, f := range found {
		out = append(out, f...)
	}
	return out
}

// checkRegistries is the step of a real release between its preflight and its
// first publish: it says the warnings and refuses on the blocking findings.
func checkRegistries(s *session.Session, plan Plan, probe *RegistryProbe, out io.Writer) error {
	targets := RegistryTargets(s, plan)
	if len(targets) == 0 {
		return nil
	}
	if probe == nil {
		probe = ProbeFromEnvironment()
		probe.NuGetTrusted = s.Config.Release.NuGetTrustedOn()
	}
	if probe.Published == nil {
		probe.Published = func(project, _, name string) bool {
			j, err := ReadJournal(s.Root, project)
			if err != nil {
				return false
			}
			for _, t := range targets {
				if t.Project == project && JournalHas(j, name, t.Version) {
					return true
				}
			}
			return false
		}
	}
	var blockers []string
	for _, f := range ValidateRegistries(targets, probe) {
		if f.Blocking {
			blockers = append(blockers, f.String())
		} else {
			fmt.Fprintf(out, "warning: %s\n", f)
		}
	}
	if len(blockers) > 0 {
		return fmt.Errorf("cannot release: a registry would refuse it, so nothing was published:\n  - %s", strings.Join(blockers, "\n  - "))
	}
	return nil
}

// journalRelease records what a package's release did, whether it finished
// or stopped; a journal that cannot be written is said, never fatal.
func journalRelease(s *session.Session, res ReportResult, out io.Writer) {
	if err := AppendJournal(s.Root, res.Project, JournalEntries(res, time.Now().UTC())); err != nil {
		fmt.Fprintf(out, "%s: could not write the release journal: %v\n", res.Project, err)
	}
}
