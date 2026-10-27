package doctor

import (
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
	out := []protocol.Finding{boundaryFinding("projects depend only on what the boundaries allow", errs, false)}
	if len(warns) > 0 {
		out = append(out, boundaryFinding("no project depends on an app", warns, true))
	}
	return out
}

func boundaryFinding(check string, vs []boundaries.Violation, warning bool) protocol.Finding {
	f := protocol.Finding{Check: check, OK: len(vs) == 0, Warning: warning}
	if len(vs) == 0 {
		return f
	}
	edges := make([]string, len(vs))
	for i, v := range vs {
		edges[i] = v.Edge() + ": " + v.Rule
	}
	f.Detail = Summary(edges, "; ", shownProblems)
	f.Remedy = vs[0].Hint
	if warning {
		f.Remedy += "; set kinds = true under [boundaries] in smith.toml to make this fail smith ci verify, or kinds = false to switch the rule off"
	}
	f.Remedy += "\nsmith boundaries lists every dependency that breaks a rule"
	return f
}
