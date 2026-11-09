package doctor

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/doctorchecks"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/architecture"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// conventionCoverageChecks warns when an architecture convention smith.toml
// turns on skips projects because it does not apply to their language. The
// convention's own plugin only ever sees the projects it applies to, so only
// smith can say which were left out. Only an architecture is meant to cover
// every project: a tooling convention (playwright, docker) applies where its
// tool is, so the projects it leaves out are no finding.
func conventionCoverageChecks(s *session.Session) []protocol.Finding {
	var out []protocol.Finding
	for _, c := range s.Plugins.Clients {
		name := c.Info.Plugin.Name
		conv, configured := s.Config.Conventions[name]
		if !configured || c.Info.Kind != protocol.KindConvention || !isArchitecture(name) {
			continue
		}
		cov := projectgraph.ConventionCoverage(s.Graph.Projects, c.Info.Languages, conv)
		if f, ok := CoverageFinding(name, c.Info.Languages, cov); ok {
			out = append(out, f)
		}
	}
	return out
}

// CoverageFinding is the warning for a convention that skips projects by
// language; ok is false when it skips none.
func CoverageFinding(name string, declared []string, cov projectgraph.Coverage) (protocol.Finding, bool) {
	if len(cov.Skipped) == 0 {
		return protocol.Finding{}, false
	}
	left := make([]string, len(cov.Skipped))
	for i, p := range cov.Skipped {
		left[i] = fmt.Sprintf("%s (%s)", p.Name, p.Language)
	}
	sort.Strings(left)
	f := protocol.Finding{Warning: true, Area: doctorchecks.AreaLint}
	list := doctorchecks.Summary(left, ", ", doctorchecks.ShownProblems)
	applies := "it applies to " + strings.Join(declared, ", ")
	if len(cov.Checked) == 0 {
		f.Check = name + " checks at least one project"
		f.Problem = name + " checks none of your projects"
		f.Detail = name + " checks none of your projects: " + applies + ", and your projects are " + list
	} else {
		f.Check = name + " covers every project"
		f.Problem = name + " does not check " + list
		f.Detail = name + " does not check " + list + ": " + applies
	}
	f.Remedy = "if that is on purpose, name them in exclude = [...] under [conventions." + name + "] in smith.toml; otherwise choose a convention for their language (smith architecture lists the choices and what each applies to)"
	return f, true
}

// isArchitecture reports whether the convention name is one of the
// architectures smith offers (smith architecture lists them).
func isArchitecture(name string) bool {
	_, err := architecture.Find(name)
	return err == nil && name != architecture.None
}
