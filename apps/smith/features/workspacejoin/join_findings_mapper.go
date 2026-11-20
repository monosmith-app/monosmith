package workspacejoin

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckSuffix ends the check of every join finding: "<folder> is joined to
// its language's workspace".
const CheckSuffix = " is joined to its language's workspace"

// IsJoinCheck says whether a doctor check is a join check.
//
// Parameters: check is a finding's Check.
// Returns: true for a join finding.
func IsJoinCheck(check string) bool { return strings.HasSuffix(check, CheckSuffix) }

// Findings turns a plugin's join answer into doctor findings: one for each
// project the plugin reported, passed when it is joined. A failed one carries
// the plugin's fix, and its remedy also names the way to keep the project out
// of the workspace on purpose.
//
// Parameters: projects are the projects the plugin was asked about; res is
// its answer.
// Returns: the findings, in the order of the answer.
func Findings(projects []protocol.Project, res protocol.JoinResult) []protocol.Finding {
	roots := map[string]string{}
	for _, p := range projects {
		roots[p.Name] = p.Root
	}
	var out []protocol.Finding
	for _, s := range res.Projects {
		root, ok := roots[s.Project]
		if !ok {
			continue
		}
		f := protocol.Finding{Check: root + CheckSuffix, Project: s.Project, OK: s.Joined}
		if !s.Joined {
			f.Problem = s.Problem
			if f.Problem == "" {
				f.Problem = "is not joined to " + s.Workspace
			}
			f.Detail = "smith finds " + root + " by its manifest, but it is missing from " + s.Workspace
			remedy := "join it by hand: add " + root + " to " + s.Workspace
			if s.Fix != nil {
				remedy = "smith doctor --fix: " + s.Fix.Description
			}
			f.Remedy = remedy + "\nor, if " + root + " stays out of the workspace on purpose: add `exclude = [\"" + root + "\"]` under [workspace] in smith.toml"
			f.Fix = s.Fix
		}
		out = append(out, f)
	}
	return out
}

// UnreachedLine says which projects an install did not reach and the step
// that joins them, for the install's log.
//
// Parameters: names are the projects the plugin did not reach.
// Returns: the line, or "" when names is empty.
func UnreachedLine(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return "not installed: " + strings.Join(names, ", ") + " (not joined to its language's workspace)\n  → smith doctor --fix joins it, then run smith install again"
}
