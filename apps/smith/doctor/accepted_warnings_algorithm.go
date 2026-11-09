package doctor

import (
	"github.com/monosmith-app/monosmith/apps/smith/doctorchecks"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// applyAccepted turns each warning the workspace accepts ([[doctor.accept]])
// into a passed check that keeps the reason, so the report shows it was
// weighed and not lost. An entry that matches no warning becomes a warning
// itself, so the list does not outlive the problems it answered. Only a
// warning can be accepted: a failure still fails.
func applyAccepted(findings []Finding, accept []workspace.AcceptedWarning) []Finding {
	used := make([]bool, len(accept))
	for i, f := range findings {
		if f.OK || !f.Warning {
			continue
		}
		for j, a := range accept {
			if !accepts(a, f) {
				continue
			}
			used[j] = true
			findings[i].Finding = protocol.Finding{Check: f.Check, OK: true, Project: f.Project, Area: f.Area, Detail: "Accepted: " + a.Reason}
			break
		}
	}
	for j, a := range accept {
		if used[j] {
			continue
		}
		findings = append(findings, Finding{Source: "smith", Finding: protocol.Finding{
			Check:   "accepted warning \"" + a.Check + "\" still applies",
			Problem: "smith.toml accepts the warning \"" + a.Check + "\", and doctor raises no such warning",
			Area:    doctorchecks.AreaConfig,
			Warning: true,
			Remedy:  "remove this [[doctor.accept]] entry from smith.toml",
		}})
	}
	return findings
}

// accepts says whether the entry names this warning, by its check or its
// problem, in the project the entry limits it to.
func accepts(a workspace.AcceptedWarning, f Finding) bool {
	if a.Project != "" && a.Project != f.Project {
		return false
	}
	return a.Check == f.Check || a.Check == f.Problem
}
