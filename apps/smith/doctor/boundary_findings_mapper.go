package doctor

import (
	"github.com/monosmith-app/monosmith/apps/smith/doctorchecks"
	"strconv"

	"github.com/monosmith-app/monosmith/apps/smith/boundaries"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// boundaryChecks reports the dependencies between projects that break a rule
// of smith.toml [boundaries]: one failed finding for the errors and one
// warning for what the built-in rule found while smith.toml has not turned it
// on. A workspace with none has one passed finding.
func boundaryChecks(s *session.Session) []protocol.Finding {
	return BoundaryFindings(boundaries.Check(s.Graph, s.Config.Boundaries))
}

// The checks of the boundary findings: the rules of smith.toml, and the
// built-in rule that no project depends on an app.
const (
	BoundaryRulesCheck = "projects depend only on what the boundaries allow"
	BoundaryAppCheck   = "no project depends on an app"
)

// BoundaryFindings maps violations to findings.
func BoundaryFindings(vs []boundaries.Violation) []protocol.Finding {
	var errs, warns []boundaries.Violation
	for _, v := range vs {
		if v.Warning {
			warns = append(warns, v)
		} else {
			errs = append(errs, v)
		}
	}
	out := []protocol.Finding{boundaryFinding(BoundaryRulesCheck, errs, false)}
	if len(warns) > 0 {
		out = append(out, boundaryFinding(BoundaryAppCheck, warns, true))
	}
	return out
}

func boundaryFinding(check string, vs []boundaries.Violation, warning bool) protocol.Finding {
	f := protocol.Finding{Check: check, OK: len(vs) == 0, Warning: warning, Area: doctorchecks.AreaDependencies}
	if len(vs) == 0 {
		return f
	}
	edges := make([]string, len(vs))
	for i, v := range vs {
		edges[i] = v.Edge() + ": " + v.Rule
	}
	f.Problem = edges[0]
	if len(edges) > 1 {
		f.Problem = strconv.Itoa(len(edges)) + " dependencies break a boundary rule"
		if warning {
			f.Problem = strconv.Itoa(len(edges)) + " dependencies point at an app"
		}
	}
	f.Detail = doctorchecks.Summary(edges, "; ", doctorchecks.ShownProblems)
	f.Remedy = vs[0].Hint
	if warning {
		f.Remedy += "; set kinds = true under [boundaries] in smith.toml to make this fail smith ci verify, or kinds = false to switch the rule off"
	}
	f.Remedy += "\nsmith boundaries lists every dependency that breaks a rule"
	return f
}
