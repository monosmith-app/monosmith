package release

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// RegistryTargets lists the registries a plan publishes an app's binaries
// to, in publishing order: npm, PyPI, NuGet, RubyGems, per package.
func RegistryTargets(s *session.Session, plan Plan) []RegistryTarget {
	var out []RegistryTarget
	for _, item := range plan.Items {
		p := s.Graph.Get(item.Project)
		if p == nil || p.Release == nil || Native(p) || s.Git.TagCommit(item.Tag) != "" {
			continue // a tagged package is finished: only its GitHub Release resumes
		}
		r := p.Release
		add := func(registry string, names ...string) {
			out = append(out, RegistryTarget{Project: item.Project, Registry: registry, Names: names, Version: item.Next})
		}
		if len(r.NPM) > 0 {
			add("npm", r.NPM...)
		}
		if r.PyPI != "" {
			add("PyPI", r.PyPI)
		}
		if r.NuGet != "" {
			add("NuGet", r.NuGet)
		}
		if r.RubyGems != "" {
			add("RubyGems", r.RubyGems)
		}
	}
	return out
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
